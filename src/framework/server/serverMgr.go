package server

import (
	"fmt"
	"framework"
	"framework/response"
	"golang.org/x/net/http2"
	"golang.org/x/net/websocket"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	controllerMap             map[string]Controller
	staticFileMap             map[string]string
	webSocketControllerMap    map[string]WebSocketController
	childHandlerControllerMap map[string]Controller
	port                      int
	staticFileMapLock         sync.Mutex
}

var serverMgrInstance *serverMgr = nil
var serverMgrOnce sync.Once

func ShareServerMgrInstance() *serverMgr {
	serverMgrOnce.Do(func() {
		serverMgrInstance = &serverMgr{}
		serverMgrInstance.controllerMap = nil
		serverMgrInstance.staticFileMap = nil
		serverMgrInstance.port = defaultServerPort
	})
	return serverMgrInstance
}

func (s *serverMgr) RegisterController(controller Controller) {
	registerController := func(controllerMap *map[string]Controller, path interface{},
		controller Controller) {
		switch path.(type) {
		case string:
			if _, ok := (*controllerMap)[path.(string)]; ok {
				fmt.Println("controller has been registered!")
				return
			}
			(*controllerMap)[path.(string)] = controller
		case []string:
			for _, p := range path.([]string) {
				if _, ok := (*controllerMap)[p]; ok {
					fmt.Println("controller has been registered!")
					return
				}
				(*controllerMap)[p] = controller
			}
		}
	}
	if normalController, ok := controller.(NormalController); ok {
		if s.controllerMap == nil {
			s.controllerMap = make(map[string]Controller)
		}
		registerController(&s.controllerMap, normalController.Path(), normalController)
	} else if childHandlerController, ok := controller.(ChildHandlerController); ok {
		if s.childHandlerControllerMap == nil {
			s.childHandlerControllerMap = make(map[string]Controller)
		}
		path, enableChildPath := childHandlerController.Path()
		registerController(&s.controllerMap, path, childHandlerController)
		if enableChildPath {
			registerController(&s.childHandlerControllerMap, path, childHandlerController)
		}
	}
}

func (s *serverMgr) RegisterWebSocketController(controller WebSocketController) {
	if s.webSocketControllerMap == nil {
		s.webSocketControllerMap = make(map[string]WebSocketController)
	}
	if path, ok := controller.Path().(string); ok {
		s.webSocketControllerMap[path] = controller
	}
	if pathList, ok := controller.Path().([]string); ok {
		for _, path := range pathList {
			s.webSocketControllerMap[path] = controller
		}
	}
}

func (s *serverMgr) RegisterStaticFile(webPath string, localPath string) {
	s.staticFileMapLock.Lock()
	defer s.staticFileMapLock.Unlock()
	if s.staticFileMap == nil {
		s.staticFileMap = make(map[string]string)
	}
	if _, ok := s.staticFileMap[webPath]; ok {
		fmt.Println("static file has beed registered!")
		return
	}
	// walkPath := filepath.Join(localPath, webPath)
	walkPath := localPath
	fmt.Println("walkPath: ", walkPath)
	filepath.Walk(walkPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fmt.Println("path: ", path)
			rel, _ := filepath.Rel(localPath, path)
			fmt.Println(rel)
			webFilePath := filepath.Join(webPath, rel)
			//webFilePath := path[len(localPath)+1:]
			s.staticFileMap[webFilePath], err = filepath.Abs(path)
			fmt.Println(webFilePath)
			fmt.Println(s.staticFileMap[webFilePath])
		}
		return nil
	})
}

func (s *serverMgr) UnRegisterStaticFile(webPath string, localPath string) {
	s.staticFileMapLock.Lock()
	defer s.staticFileMapLock.Unlock()
}

func (s *serverMgr) SetServerPort(port int) {
	s.port = port
}

func (s *serverMgr) handlerWebsocketReq(w http.ResponseWriter, r *http.Request) bool {
	if controller, ok := s.webSocketControllerMap[r.URL.Path]; ok {
		websocket.Handler(controller.HandlerRequest).ServeHTTP(w, r)
		return true
	}
	return false
}

func (s *serverMgr) handlerStatisFileReq(w http.ResponseWriter, currentPath string) bool {
	s.staticFileMapLock.Lock()
	defer s.staticFileMapLock.Unlock()
	if currentPath[0] == '/' {
		currentPath = currentPath[1:]
	}
	if local, ok := s.staticFileMap[currentPath]; ok {
		ext := filepath.Ext(local)
		contentType := ""
		if v, ok := extContentTypeMap[strings.ToLower(ext)]; ok {
			contentType = v
		} else {
			contentType = "application/octet-stream"
		}
		file, err := os.Open(local)
		if err != nil {
			response.JsonResponseWithMsg(w, framework.ErrorNoSuchFileOrDirectory, err.Error())
			return false
		}
		defer file.Close()
		fileInfo, err := os.Stat(local)
		if err != nil {
			response.JsonResponseWithMsg(w, framework.ErrorNoSuchFileOrDirectory, err.Error())
			return false
		}
		content := make([]byte, fileInfo.Size())
		file.Read(content)
		w.Header().Set("Accept", "*/*")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		w.Header().Set("Content-Type", contentType)
		w.Write(content)
		return true
	}
	return false
}

func (s *serverMgr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	currentPath := r.URL.Path
	// 1. 首先在controller里面寻找
	if controller, ok := s.controllerMap[currentPath]; ok {
		controller.HandlerRequest(w, r)
		return
	}
	// 2. 在static file 里面寻找
	if s.handlerStatisFileReq(w, currentPath) {
		return
	}
	// 3. 逐级分解，看是不是某个controller的子集
	for true {
		lastIndex := strings.LastIndex(currentPath, "/")
		if lastIndex != -1 {
			currentPath = currentPath[:lastIndex]
			if controller, ok := s.childHandlerControllerMap[currentPath]; ok {
				controller.HandlerRequest(w, r)
				return
			}
		} else {
			break
		}
	}
	// 4. websocket
	if s.handlerWebsocketReq(w, r) {
		return
	}
	// 5. 404
	fmt.Println("404: ", r.URL.Path)
	w.WriteHeader(http.StatusNotFound)
}

func (s *serverMgr) startHTTPServer() {
	http.HandleFunc("/", s.ServeHTTP)
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *serverMgr) startHTTPSServer() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s,
	}
	http2.ConfigureServer(srv, &http2.Server{})
	fmt.Println(srv.ListenAndServeTLS("cert.pem", "key.pem"))
}

func (s *serverMgr) StartServer() {
	s.startHTTPServer()
}
