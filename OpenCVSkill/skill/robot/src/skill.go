package examples

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"mind/core/framework"
	"mind/core/framework/drivers/hexabody"
	"mind/core/framework/drivers/media"  //for the camera
	"mind/core/framework/log"
	"mind/core/framework/skill" //for skill.Base
	"os"

	"github.com/lazywei/go-opencv/opencv"
)

//the struct called "OpenCVSkill"
//stuff needed in the skill is everything under the "skill.Base" thing
type OpenCVSkill struct {
	skill.Base
	stop    chan bool  //field#1 is a "bool" channel
	cascade *opencv.HaarCascade   //field#2 is a pointer to "opencv.HaarCascade"
}


//the NewSkill() function allows the mind SDK to see that "OpenCVSkill" is actually
//a struct that will be able to implement all of the methods inside the "skill.Base" interface

//returns the initialized "OpenCVSkill" struct to the "codes inside HEXA" for HEXA to manipulate
//the data inside this "OpenCVSkill" struct using methods
func NewSkill() skill.Interface {
	return &OpenCVSkill{
		stop:    make(chan bool),
		cascade: opencv.LoadHaarClassifierCascade("assets/haarcascade_frontalface_alt.xml"),
	}
}

//define the method on the "OpenCVSkill"
//this method is user-defined!!
func (d *OpenCVSkill) sight() {
	for {
		select {    //only when other cases are not ready, the default case will run
		case <-d.stop:
			return
		default:  //this case will run if the "<-d.stop" is not ready
			image := media.SnapshotRGBA()  //take a picture and stores it into image

			//start encoding the "image" picture just taken and sent it to "remote"
			buf := new(bytes.Buffer)
			jpeg.Encode(buf, image, nil)
			str := base64.StdEncoding.EncodeToString(buf.Bytes())
			framework.SendString(str)

			//reference: https://github.com/go-opencv/go-opencv/blob/master/opencv/goimage.go
			//"FromImageUnsafe" will create a buffer for the "cvimg" to share the same buffer as the go's "image"
			cvimg := opencv.FromImageUnsafe(image)

			//reference: https://github.com/go-opencv/go-opencv/blob/master/opencv/cvaux.go
			//returns an array that tells you what range the face is in?
			faces := d.cascade.DetectObjects(cvimg)
			log.Info.Print("The length of the faces are: "); log.Info.Println(len(faces));
			hexabody.StandWithHeight(float64(len(faces)) * 50)
		}
	}
}


//when the skill is started
func (d *OpenCVSkill) OnStart() {
	log.Info.Println("Started")
}


//when HEXA is connected to the remote = laptops
func (d *OpenCVSkill) OnConnect() {
	//make hexa stand and check if there is errors
	err := hexabody.Start()
	if err != nil {
		log.Error.Println("Hexabody start err:", err)
		return
	}
	if !media.Available() {
		log.Error.Println("Media driver not available")
		return
	}
	if err := media.Start(); err != nil {
		log.Error.Println("Media driver could not start")
	}
}



//when the skill is shutting down
func (d *OpenCVSkill) OnClose() {
	hexabody.Close()
}



func (d *OpenCVSkill) OnDisconnect() {
	log.Info.Println("Disconnecting from the remote")
	os.Exit(0) // Closes the process when remote disconnects
}



func (d *OpenCVSkill) OnRecvString(data string) {
	log.Info.Println(data)
	switch data {
	case "start":
		go d.sight()           //start the skill, automatically falls into the "default" part of the selection in "sight()"
	case "stop":
		d.stop <- true      //when the "stop" button is pressed, the "sight()" function will be stopped
	}
}
