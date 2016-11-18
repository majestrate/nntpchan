package frontend

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/majestrate/srndv2/lib/config"
	"net/http"
)

// server of captchas
// implements frontend.Middleware
type CaptchaServer struct {
	h           int
	w           int
	store       *sessions.CookieStore
	prefix      string
	sessionName string
}

// create new captcha server using existing session store
func NewCaptchaServer(w, h int, prefix string, store *sessions.CookieStore) *CaptchaServer {
	return &CaptchaServer{
		h:           h,
		w:           w,
		prefix:      prefix,
		store:       store,
		sessionName: "captcha",
	}
}

func (cs *CaptchaServer) Reload(c *config.MiddlewareConfig) {

}

func (cs *CaptchaServer) SetupRoutes(m *mux.Router) {
	m.Path("/new").HandlerFunc(cs.NewCaptcha)
	m.Path("/img/{f}").Handler(captcha.Server(cs.w, cs.h))
	m.Path("/verify.json").HandlerFunc(cs.VerifyCaptcha)
}

// return true if this session has solved the last captcha given provided solution, otherwise false
func (cs *CaptchaServer) CheckSession(w http.ResponseWriter, r *http.Request, solution string) (bool, error) {
	s, err := cs.store.Get(r, cs.sessionName)
	if err == nil {
		id, ok := s.Values["captcha_id"]
		if ok {
			return captcha.VerifyString(id.(string), solution), nil
		}
	}
	return false, err
}

// verify a captcha
func (cs *CaptchaServer) VerifyCaptcha(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	// request
	req := make(map[string]string)
	// response
	resp := make(map[string]interface{})
	resp["solved"] = false
	// decode request
	err := dec.Decode(req)
	if err == nil {
		// decode okay
		id, ok := req["id"]
		if ok {
			// we have id
			solution, ok := req["solution"]
			if ok {
				// we have solution and id
				resp["solved"] = captcha.VerifyString(id, solution)
			} else {
				// we don't have solution
				err = errors.New("no captcha solution provided")
			}
		} else {
			// we don't have id
			err = errors.New("no captcha id provided")
		}
	}
	if err != nil {
		// error happened
		resp["error"] = err.Error()
	}
	// send reply
	w.Header().Set("Content-Type", "text/json; encoding=UTF-8")
	enc := json.NewEncoder(w)
	enc.Encode(resp)
}

// generate a new captcha
func (cs *CaptchaServer) NewCaptcha(w http.ResponseWriter, r *http.Request) {
	// obtain session
	sess, err := cs.store.Get(r, cs.sessionName)
	if err != nil {
		// failed to obtain session
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// new captcha
	id := captcha.New()
	// do we want to interpret as json?
	use_json := r.URL.Query().Get("t") == "json"
	// image url
	url := fmt.Sprintf("%simg/%s.png", cs.prefix, id)
	if use_json {
		// send json
		enc := json.NewEncoder(w)
		enc.Encode(map[string]string{"id": id, "url": url})
	} else {
		// set captcha id
		sess.Values["captcha_id"] = id
		// save session
		sess.Save(r, w)
		// rediect to image
		http.Redirect(w, r, url, http.StatusFound)
	}
}
