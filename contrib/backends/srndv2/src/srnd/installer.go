package srnd

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/majestrate/configparser"
	"golang.org/x/text/language"
	"gopkg.in/tylerb/graceful.v1"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type handlePost func(*dialogNode, url.Values, *configparser.Configuration) (*dialogNode, error)
type templateModel map[string]interface{}
type prepareModel func(*dialogNode, error, *configparser.Configuration) templateModel

type dialogNode struct {
	parent   *dialogNode
	children map[string]*dialogNode

	post  handlePost
	model prepareModel

	templateName string
}

type Installer struct {
	root            *dialogNode
	currentNode     *dialogNode
	currentErr      error
	result          chan *configparser.Configuration
	config          *configparser.Configuration
	srv             *graceful.Server
	hasTranslations bool
}

func handleDBTypePost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	db := form.Get("db")
	log.Println("DB chosen: ", db)
	if db == "postgres" {
		return self.children["postgres"], nil
	}
	return self, nil
}

func prepareDefaultModel(self *dialogNode, err error, conf *configparser.Configuration) templateModel {
	param := make(map[string]interface{})
	param["dialog"] = &BaseDialogModel{ErrorModel{err}, StepModel{self}}
	return param
}

func preparePostgresDBModel(self *dialogNode, err error, conf *configparser.Configuration) templateModel {
	param := make(map[string]interface{})
	sect, _ := conf.Section("database")
	host := sect.ValueOf("host")
	port := sect.ValueOf("port")
	user := sect.ValueOf("user")
	param["dialog"] = &DBModel{ErrorModel{err}, StepModel{self}, user, host, port}
	return param
}

func handlePostgresDBPost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	if form.Get("back") == "true" {
		return self.parent, nil
	}
	sect, _ := conf.Section("database")
	host := form.Get("host")
	port := form.Get("port")
	passwd := form.Get("password")
	user := form.Get("user")

	err := checkPostgresConnection(host, port, user, passwd)
	if err != nil {
		return self, err
	}
	sect.Add("type", "postgres")
	sect.Add("schema", "srnd")
	sect.Add("host", host)
	sect.Add("port", port)
	sect.Add("password", passwd)
	sect.Add("user", user)

	return self.children["next"], nil
}

func prepareNNTPModel(self *dialogNode, err error, conf *configparser.Configuration) templateModel {
	param := make(map[string]interface{})
	sect, _ := conf.Section("nntp")
	name := sect.ValueOf("instance_name")
	param["dialog"] = &NameModel{ErrorModel{err}, StepModel{self}, name}
	return param
}

func handleNNTPPost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	if form.Get("back") == "true" {
		return self.parent, nil
	}
	sect, _ := conf.Section("nntp")
	name := form.Get("nntp_name")

	allow_attachments := form.Get("allow_attachments")
	if allow_attachments != "1" {
		allow_attachments = "0"
	}

	allow_anon := form.Get("allow_anon")
	if allow_anon != "1" {
		allow_anon = "0"
	}

	allow_anon_attachments := form.Get("allow_anon_attachments")
	if allow_anon_attachments != "1" {
		allow_anon_attachments = "0"
	}

	require_tls := form.Get("require_tls")
	if require_tls != "1" {
		require_tls = "0"
	}

	sect.Add("instance_name", name)
	sect.Add("allow_attachments", allow_attachments)
	sect.Add("allow_anon", allow_anon)
	sect.Add("require_tls", require_tls)

	return self.children["next"], nil
}

func handleCryptoPost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	if form.Get("back") == "true" {
		return self.parent, nil
	}
	sect, _ := conf.Section("crypto")
	host := form.Get("host")
	key := form.Get("key")

	err := checkHost(host)
	if err != nil {
		return self, err
	}
	sect.Add("tls-hostname", host)
	sect.Add("tls-keyname", key)

	return self.children["next"], nil
}

func prepareCryptoModel(self *dialogNode, err error, conf *configparser.Configuration) templateModel {
	param := make(map[string]interface{})
	sect, _ := conf.Section("crypto")
	host := sect.ValueOf("tls-hostname")
	key := sect.ValueOf("tls-keyname")
	param["dialog"] = &CryptoModel{ErrorModel{err}, StepModel{self}, host, key}
	return param
}

func prepareBinModel(self *dialogNode, err error, conf *configparser.Configuration) templateModel {
	param := make(map[string]interface{})
	sect, _ := conf.Section("articles")
	convert := sect.ValueOf("convert_bin")
	ffmpeg := sect.ValueOf("ffmpegthumbnailer_bin")
	sox := sect.ValueOf("sox_bin")
	param["dialog"] = &BinaryModel{ErrorModel{err}, StepModel{self}, convert, ffmpeg, sox}
	return param
}

func handleBinPost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	if form.Get("back") == "true" {
		return self.parent, nil
	}
	sect, _ := conf.Section("articles")
	convert := form.Get("convert")
	ffmpeg := form.Get("ffmpeg")
	sox := form.Get("sox")

	err := checkFile(convert)
	if err == nil {
		err = checkFile(ffmpeg)
		if err == nil {
			err = checkFile(sox)
		}
	}

	sect.Add("convert_bin", convert)
	sect.Add("ffmpegthumbnailer_bin", ffmpeg)
	sect.Add("sox_bin", sox)

	if err != nil {
		return self, err
	}

	return self.children["next"], nil
}

func handleCacheTypePost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	if form.Get("back") == "true" {
		return self.parent, nil
	}
	sect, _ := conf.Section("cache")

	cache := form.Get("cache")
	log.Println("Cache chosen: ", cache)
	sect.Add("type", cache)
	if cache == "file" || cache == "null" || cache == "varnish" {
		return self.children["next"], nil
	}

	return self, nil
}

func prepareFrontendModel(self *dialogNode, err error, conf *configparser.Configuration) templateModel {
	param := make(map[string]interface{})
	sect, _ := conf.Section("frontend")
	name := sect.ValueOf("name")
	locale := sect.ValueOf("locale")
	param["dialog"] = &FrontendModel{ErrorModel{err}, StepModel{self}, name, locale}
	return param
}

func handleFrontendPost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	if form.Get("back") == "true" {
		return self.parent, nil
	}
	var next *dialogNode

	sect, _ := conf.Section("frontend")
	name := form.Get("name")
	locale := form.Get("locale")

	allow_files := form.Get("allow_files")
	if allow_files != "1" {
		allow_files = "0"
	}

	json_api := form.Get("json")
	if json_api != "1" {
		json_api = "0"
		next = self.children["next"]
	} else {
		next = self.children["json"]
	}

	sect.Add("name", name)
	sect.Add("locale", locale)
	sect.Add("allow_files", allow_files)
	sect.Add("json-api", json_api)

	err := checkLocale(locale)
	if err != nil {
		return self, err
	}

	return next, nil
}

func handleAPIPost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	if form.Get("back") == "true" {
		return self.parent, nil
	}
	sect, _ := conf.Section("frontend")
	user := form.Get("user")
	pass := form.Get("pass")
	secret := form.Get("secret")

	sect.Add("json-api-username", user)
	sect.Add("json-api-password", pass)
	sect.Add("api-secret", secret)

	return self.children["next"], nil
}

func prepareAPIModel(self *dialogNode, err error, conf *configparser.Configuration) templateModel {
	param := make(map[string]interface{})
	sect, _ := conf.Section("frontend")
	user := sect.ValueOf("json-api-username")
	secret := sect.ValueOf("api-secret")
	param["dialog"] = &APIModel{ErrorModel{err}, StepModel{self}, user, secret}
	return param
}

func handleKeyPost(self *dialogNode, form url.Values, conf *configparser.Configuration) (*dialogNode, error) {
	if form.Get("back") == "true" {
		return self.parent, nil
	}
	sect, _ := conf.Section("frontend")
	public := form.Get("public")

	sect.Add("admin_key", public)
	return self.children["next"], nil
}

func prepareKeyModel(self *dialogNode, err error, conf *configparser.Configuration) templateModel {
	param := make(map[string]interface{})
	public, secret := newSignKeypair()
	param["dialog"] = &KeyModel{ErrorModel{err}, StepModel{self}, public, secret}
	return param
}

func (self *Installer) HandleInstallerGet(wr http.ResponseWriter, r *http.Request) {
	if !self.hasTranslations {
		t, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
		locale := ""
		if len(t) > 0 {
			locale = t[0].String()
		}
		InitI18n(locale, filepath.Join("contrib", "translations"))
		self.hasTranslations = true
	}
	if self.currentNode == nil {
		wr.WriteHeader(404)
	} else {
		m := self.currentNode.model(self.currentNode, self.currentErr, self.config)
		template.writeTemplate(self.currentNode.templateName, m, wr)
	}
}

func (self *Installer) HandleInstallerPost(wr http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err == nil {
		next, newErr := self.currentNode.post(self.currentNode, r.PostForm, self.config)
		if next == nil {
			self.result <- self.config
			//defer self.srv.Stop(10 * time.Second)
		}
		self.currentNode = next
		self.currentErr = newErr

		http.Redirect(wr, r, r.URL.String(), http.StatusSeeOther) //redirect to the same url, but with a GET
		return
	}
	http.Error(wr, "Bad Request", http.StatusBadRequest)
}

func NewInstaller(result chan *configparser.Configuration) *Installer {
	inst := new(Installer)
	inst.root = initInstallerTree()
	inst.currentNode = inst.root
	inst.result = result
	inst.config = GenSRNdConfig()
	inst.hasTranslations = false

	m := mux.NewRouter()
	m.Path("/").HandlerFunc(inst.HandleInstallerGet).Methods("GET")
	m.Path("/").HandlerFunc(inst.HandleInstallerPost).Methods("POST")

	inst.srv = &graceful.Server{
		Timeout:          10 * time.Second,
		NoSignalHandling: true,

		Server: &http.Server{
			Addr:    ":18000",
			Handler: m,
		},
	}

	return inst
}

func initInstallerTree() *dialogNode {
	root := &dialogNode{
		parent:       nil,
		children:     make(map[string]*dialogNode),
		post:         handleDBTypePost,
		model:        prepareDefaultModel,
		templateName: "inst_db.mustache",
	}

	postgresDB := &dialogNode{
		parent:       root,
		children:     make(map[string]*dialogNode),
		post:         handlePostgresDBPost,
		model:        preparePostgresDBModel,
		templateName: "inst_postgres_db.mustache",
	}
	root.children["postgres"] = postgresDB

	nntp := &dialogNode{
		parent:       root,
		children:     make(map[string]*dialogNode),
		post:         handleNNTPPost,
		model:        prepareNNTPModel,
		templateName: "inst_nntp.mustache",
	}
	postgresDB.children["next"] = nntp

	crypto := &dialogNode{
		parent:       nntp,
		children:     make(map[string]*dialogNode),
		post:         handleCryptoPost,
		model:        prepareCryptoModel,
		templateName: "inst_crypto.mustache",
	}
	nntp.children["next"] = crypto

	bins := &dialogNode{
		parent:       crypto,
		children:     make(map[string]*dialogNode),
		post:         handleBinPost,
		model:        prepareBinModel,
		templateName: "inst_bins.mustache",
	}
	crypto.children["next"] = bins

	cache := &dialogNode{
		parent:       bins,
		children:     make(map[string]*dialogNode),
		post:         handleCacheTypePost,
		model:        prepareDefaultModel,
		templateName: "inst_cache.mustache",
	}
	bins.children["next"] = cache

	frontend := &dialogNode{
		parent:       cache,
		children:     make(map[string]*dialogNode),
		post:         handleFrontendPost,
		model:        prepareFrontendModel,
		templateName: "inst_frontend.mustache",
	}
	cache.children["next"] = frontend

	api := &dialogNode{
		parent:       frontend,
		children:     make(map[string]*dialogNode),
		post:         handleAPIPost,
		model:        prepareAPIModel,
		templateName: "inst_api.mustache",
	}
	frontend.children["json"] = api

	key := &dialogNode{
		parent:       frontend,
		children:     make(map[string]*dialogNode),
		post:         handleKeyPost,
		model:        prepareKeyModel,
		templateName: "inst_key.mustache",
	}
	frontend.children["next"] = key
	api.children["next"] = key

	return root
}

func checkPostgresConnection(host, port, user, password string) error {
	var db_str string
	if len(user) > 0 {
		if len(password) > 0 {
			db_str = fmt.Sprintf("user=%s password=%s host=%s port=%s client_encoding='UTF8' connect_timeout=3", user, password, host, port)
		} else {
			db_str = fmt.Sprintf("user=%s host=%s port=%s client_encoding='UTF8' connect_timeout=3", user, host, port)
		}
	} else {
		if len(port) > 0 {
			db_str = fmt.Sprintf("host=%s port=%s client_encoding='UTF8' connect_timeout=3", host, port)
		} else {
			db_str = fmt.Sprintf("host=%s client_encoding='UTF8' connect_timeout=3", host)
		}
	}

	conn, err := sql.Open("postgres", db_str)
	defer conn.Close()

	if err == nil {
		_, err = conn.Exec("SELECT datname FROM pg_database")
	}

	return err
}

func checkLocale(locale string) error {
	_, err := language.Parse(locale)
	return err
}

func checkFile(path string) error {
	_, err := os.Stat(path)
	return err
}

func checkHost(host string) error {
	_, err := net.LookupHost(host)
	return err
}

func (self *Installer) Start() {
	log.Println("starting installer on", self.srv.Server.Addr)
	log.Println("open up http://127.0.0.1:18000 to do initial configuration")
	self.srv.ListenAndServe()
}

func (self *Installer) Stop() {
	self.srv.Stop(1 * time.Second)
}

func InstallerEnabled() bool {
	return os.Getenv("SRND_NO_INSTALLER") != "1"
}
