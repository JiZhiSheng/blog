package server

import (
	"container/list"
	"fmt"
	"net/http"
	"sync"
)

const defaultServerPort = 8080

type controllerElement struct {
	webPath    string
	controller Controller
}

type staticFileElement struct {
	webPath   string
	localPath string
}

type serverMgr struct {
	controllerList *list.List
	staticFileList *list.List
	port           int
}

var serverMgrInstance *serverMgr = nil
var serverMgrOnce sync.Once

func ShareServerMgrInstance() *serverMgr {
	serverMgrOnce.Do(func() {
		serverMgrInstance = &serverMgr{}
		serverMgrInstance.controllerList = nil
		serverMgrInstance.staticFileList = nil
		serverMgrInstance.port = defaultServerPort
	})
	return serverMgrInstance
}

func (s *serverMgr) RegisterController(path string, controller Controller) {
	if s.controllerList == nil {
		s.controllerList = list.New()
	}
	s.controllerList.PushBack(&controllerElement{path, controller})
}

func (s *serverMgr) RegisterStaticFile(webPath string, localPath string) {
	if s.staticFileList == nil {
		s.staticFileList = list.New()
	}
	s.staticFileList.PushBack(&staticFileElement{webPath, localPath})
}

func (s *serverMgr) SetServerPort(port int) {
	s.port = port
}

func (s *serverMgr) StartServer() {
	// register controller
	if s.controllerList != nil {
		for controller := s.controllerList.Front(); controller != nil; controller = controller.Next() {
			element := controller.Value.(*controllerElement)
			http.HandleFunc(element.webPath, element.controller.HandlerAction)
		}
	}

	// register static file
	if s.staticFileList != nil {
		for file := s.staticFileList.Front(); file != nil; file = file.Next() {
			element := file.Value.(*staticFileElement)
			http.Handle(element.webPath, http.FileServer(http.Dir(element.localPath)))
		}
	}

	fmt.Println("server at port: ", s.port)
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}