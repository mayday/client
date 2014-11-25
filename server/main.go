package server

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/base64"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"mayday/core"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

const (
	DefaultPort = 8080
)

type File struct {
	Id      int `orm:"auto"`
	Path    string
	Created time.Time `orm:"auto_now_add;type(datetime)"`
	Case    *Case     `orm:"rel(fk)"`
}

type UploadFile struct {
	Filename string
	Content  string
}

type Case struct {
	Id          int       `orm:"auto"`
	Description string    `orm:"default(""), type(text)"`
	Created     time.Time `orm:"auto_now_add;type(datetime)"`
	Updated     time.Time `orm:"auto_now_add;type(datetime)"`
	IsSigned    bool      `orm:"default(false)"`
	IsPrivate   bool      `orm:"default(false)"`
	Token       string
	Config      string  `orm:"default(""), type(text)"`
	Signed      string  `orm:"default(""), type(text)"`
	Files       []*File `orm:"reverse(many)"`
}

type CaseHandler struct {
	StoragePath string
}

func (handler *CaseHandler) Create(request *restful.Request, response *restful.Response) {
	c := new(Case)
	err := request.ReadEntity(c)

	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if c.IsPrivate {
		c.Token = uuid.New()
	}

	if c.Signed != "" {
		c.IsSigned = true
	}

	o := orm.NewOrm()
	o.Insert(c)

	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(c)
}

func (handler *CaseHandler) Get(request *restful.Request, response *restful.Response) {
	id, err := strconv.Atoi(request.PathParameter("case-id"))

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, "invalid provided id")
		return
	}

	c := Case{Id: id}
	o := orm.NewOrm()

	err = o.Read(&c)

	if err != nil {
		response.WriteErrorString(http.StatusNotFound, err.Error())
		return
	}

	if c.IsPrivate {
		token := request.QueryParameter("token")
		if c.Token != token || token == "" {
			response.WriteErrorString(http.StatusForbidden, "Invalid Token")
			return
		}
	}

	o.LoadRelated(&c, "Files")
	response.WriteEntity(c)
}

func (handler *CaseHandler) GetFile(request *restful.Request, response *restful.Response) {
	id, err := strconv.Atoi(request.PathParameter("case-id"))

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, "invalid provided id")
		return
	}

	c := Case{Id: id}
	o := orm.NewOrm()

	err = o.Read(&c)

	if err != nil {
		response.WriteErrorString(http.StatusNotFound, err.Error())
		return
	}

	if c.IsPrivate {
		token := request.QueryParameter("token")
		if c.Token != token || token == "" {
			response.WriteErrorString(http.StatusForbidden, "Invalid Token")
			return
		}
	}

	file_id, err := strconv.Atoi(request.PathParameter("file-id"))
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, "invalid provided file id")
		return
	}

	o.LoadRelated(&c, "Files")

	for _, file := range c.Files {
		if file.Id == file_id {
			fullpath := path.Join(handler.StoragePath, strconv.Itoa(id), file.Path)

			readed, err := ioutil.ReadFile(fullpath)
			if err != nil {
				response.WriteErrorString(http.StatusInternalServerError, "cannot read file")
				return
			}

			response.WriteEntity(UploadFile{
				Filename: file.Path,
				Content:  base64.StdEncoding.EncodeToString(readed),
			})
			return
		}
	}

	response.WriteErrorString(http.StatusNotFound, "not found specified file")
	return
}

func (handler *CaseHandler) UploadFiles(request *restful.Request, response *restful.Response) {
	id, err := strconv.Atoi(request.PathParameter("case-id"))

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, "invalid provided id")
		return
	}

	c := Case{Id: id}
	o := orm.NewOrm()

	err = o.Read(&c)

	if err != nil {
		response.WriteErrorString(http.StatusNotFound, err.Error())
		return
	}

	if c.IsPrivate {
		token := request.QueryParameter("token")
		if c.Token != token || token == "" {
			response.WriteErrorString(http.StatusForbidden, "Invalid Token")
			return
		}
	}

	f := new(UploadFile)
	err = request.ReadEntity(f)

	if err != nil {
		response.AddHeader("Content-Type", "application/json")
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	base := path.Join(handler.StoragePath, strconv.Itoa(id))
	if _, err := os.Stat(base); os.IsNotExist(err) {
		os.Mkdir(base, 0700)
	}

	data, err := base64.StdEncoding.DecodeString(f.Content)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, "Invalid file contents")
		return
	}

	output := path.Join(base, f.Filename)
	err = ioutil.WriteFile(output, data, 0700)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, "Cannot store file")
		return
	}

	new_file := &File{}
	new_file.Path = f.Filename
	new_file.Case = &c

	o.Insert(new_file)

	response.WriteHeader(http.StatusCreated)
}

func init() {
	orm.RegisterModel(new(Case), new(File))
	orm.RegisterDataBase("default", "sqlite3", "/tmp/database.db", 30)
	orm.RunCommand()

	name := "default"
	force := false
	verbose := false

	err := orm.RunSyncdb(name, force, verbose)
	if err != nil {
		fmt.Println(err)
	}
}

func GetDefaultStoragePath() string {
	base, err := core.GetDefaultDirectory()

	if err != nil {
		return ""
	}

	base = path.Join(base, "files")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		os.Mkdir(base, 0700)
	}

	return base
}

func Start(bind string, port int, storage string) {
	handler := &CaseHandler{}

	if _, err := os.Stat(storage); os.IsNotExist(err) || storage == "" {
		storage = GetDefaultStoragePath()
	}

	handler.StoragePath = storage

	ws := new(restful.WebService)
	ws.Path("/1/case").
		Doc("Manage support reports").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/{case-id}").To(handler.Get).
		Doc("get a specific report").
		Operation("findCase").
		Param(ws.PathParameter("case-id", "case identifier").DataType("int")).
		Param(ws.QueryParameter("token", "private token identifier")).
		Writes(Case{}))

	ws.Route(ws.GET("/{case-id}/file/{file-id}").To(handler.GetFile).
		Doc("get a specific file report").
		Operation("findCase").
		Param(ws.PathParameter("case-id", "case identifier").DataType("int")).
		Param(ws.PathParameter("file-id", "file identifier").DataType("int")).
		Param(ws.QueryParameter("token", "private token identifier")).
		Writes(Case{}))

	ws.Route(ws.POST("/{case-id}/file").To(handler.UploadFiles).
		Doc("Upload a file to a specific case").
		Param(ws.PathParameter("case-id", "case identifier").DataType("int")).
		Param(ws.QueryParameter("token", "private token identifier")).
		Writes(Case{}))

	ws.Route(ws.POST("").To(handler.Create).
		Doc("create a case").
		Operation("createCase").
		Reads(Case{})) // from the request

	container := restful.NewContainer()
	container.Add(ws)

	config := swagger.Config{
		WebServices:    container.RegisteredWebServices(),
		WebServicesUrl: "http://localhost:8080",
		ApiPath:        "/apidocs.json",
		SwaggerPath:    "/apidocs/",
	}

	swagger.RegisterSwaggerService(config, container)
	log.Printf("start listening on localhost:%s", port)
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: container}
	log.Fatal(server.ListenAndServe())
}
