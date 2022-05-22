package main

import (
	"fmt"
	"os"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"

	"github.com/g3n/engine/experimental/collision/shape"
	"github.com/g3n/engine/experimental/physics"
	"github.com/g3n/engine/experimental/physics/object"

	"github.com/DokaiStudio/engine/gblk"
)

type DokaiStudio struct {
	app *app.Application

	simulasi *physics.Simulation
	rootpart  *object.Body

	//attractorOn bool
	gravity     *physics.ConstantForceField
	attractor   *physics.AttractorForceField
}

func main() {
	// Create application and run
	studio := &DokaiStudio{
		app: app.App(),
	}
	runApp(studio, &os.Args)

	/*
	file, err := os.Open("tes.gblk")
	if err != nil {
		panic(err)
	}

	lexer := gblk.NewLexer(file)
	for {
		pos, tok, lit := lexer.Lex()
		if tok == gblk.EOF {
			break
		}

		fmt.Printf("%d:%d\t%s\t%s\n", pos.Line, pos.Column, tok, lit)
	}*/

}

func runApp(studio *DokaiStudio, runargs *[]string) {
	fmt.Printf("run args: %s", runargs)

    a := studio.app

	glfwWindow := a.IWindow.(*window.GlfwWindow)
	glfwWindow.SetTitle("Amogus Studio")

	scene := core.NewNode()
	studio.simulasi = physics.NewSimulation(scene)
	studio.gravity = physics.NewConstantForceField(&math32.Vector3{0, -0.98, 0})
	studio.attractor = physics.NewAttractorForceField(&math32.Vector3{0, 1, 0}, 1)
	//studio.simulasi.AddForceField(studio.gravity)

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

    // Create axes helper
    axes := helper.NewAxes(3)
	scene.Add(axes)
	axesbody := object.NewBody(axes)
	axesbody.SetShape(shape.NewPlane())
	studio.simulasi.AddBody(axesbody, "RootPartAxes")
    
	// Create perspective camera
	cam := camera.New(1)
	cam.SetPosition(0, 0, 3)
	cam.SetName("MainCamera")
	scene.Add(cam)

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		cam.SetAspect(float32(width) / float32(height))
	}
	OnExit := func(evname string, ev interface{}) {
        fmt.Printf("app exited: %dms\n", a.RunTime().Milliseconds())
	}
	a.Subscribe(window.OnWindowSize, onResize)
	a.Subscribe(app.OnExit, OnExit)
	onResize("", nil)

	// Create a box as root part
	geom := geometry.NewBox(5, .2, 5) //NewTorus(1, .4, 12, 32, math32.Pi*2)
	mat := material.NewStandard(math32.NewColor("White"))
	mesh := graphic.NewMesh(geom, mat)
	scene.Add(mesh)
	studio.rootpart = object.NewBody(mesh)
	studio.rootpart.SetShape(shape.NewPlane())
	studio.simulasi.AddBody(studio.rootpart, "RootPart")
	//studio.rootpart.SetVelocity(math32.NewVector3(-0.5, 0, 0))
	studio.rootpart.SetAngularVelocity(math32.NewVector3(0, 0.4, 0))
	axesbody.SetAngularVelocity(math32.NewVector3(0, 0.4, 0))

	//scene.Add(helper.NewNormals(mesh, 0.5, &math32.Color{0, 0, 1}, 1))

    //FPS
	txtlfps := "FPS: "  
	lfpsinfo := gui.NewLabel(txtlfps)
	lfpsinfo.SetFontSize(18)
	lfpsinfo.SetColor(math32.NewColor("Black"))
	scene.Add(lfpsinfo)

	// Create and add a button to the scene
	propwindow := gui.NewWindow(180, 300) 
	propwindow.SetTitle("Properties")
	propwindow.SetPosition(1, 40)
	propwindow.SetResizable(true)

	li1 := gui.NewVList(100, 200)
	li1.SetSize(propwindow.ContentHeight(), propwindow.ContentHeight())
	propwindow.Add(li1)

	txtlcolor := "color: "
    lcolorinfo := gui.NewLabel(txtlcolor)
	li1.Add(lcolorinfo)

	btn := gui.NewButton("Merah")
	btn.SetSize(40, 80)
	btn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		mat.SetColor(math32.NewColor("Red"))
		lcolorinfo.SetText(fmt.Sprintf("%srgb(%.2f, %.2f, %.2f)", txtlcolor, mat.AmbientColor().R, mat.AmbientColor().G, mat.AmbientColor().B))
	})

	btn2 := gui.NewButton("Abu-abu")
	btn2.SetSize(40, 80)
	btn2.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		mat.SetColor(math32.NewColor("Gray"))
		lcolorinfo.SetText(fmt.Sprintf("%srgb(%.2f, %.2f, %.2f)", txtlcolor, mat.AmbientColor().R, mat.AmbientColor().G, mat.AmbientColor().B))
	})

	li1.Add(btn)
	li1.Add(btn2)
	scene.Add(propwindow)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(0, 0, 2)
	scene.Add(pointLight)

	// Create and add an axis helper to the scene
	//scene.Add(helper.NewAxes(0.5))

	// Set background color to blue sky
	a.Gls().ClearColor(0.5294117647058824, 0.807843137254902,  0.9215686274509803, 1)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
		studio.simulasi.Step(float32(deltaTime.Seconds()))
		lfpsinfo.SetText(fmt.Sprintf("%s%dms", txtlfps, deltaTime.Milliseconds()))
	})
}
