package webhooks

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"nntpchan/lib/config"
	"nntpchan/lib/nntp"
	"nntpchan/lib/nntp/message"
	"nntpchan/lib/store"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strings"
)

// web hook implementation
type httpWebhook struct {
	conf    *config.WebhookConfig
	storage store.Storage
	hdr     *message.HeaderIO
}

func (h *httpWebhook) SentArticleVia(msgid nntp.MessageID, name string) {
	// web hooks don't care about feed state
}

// we got a new article
func (h *httpWebhook) GotArticle(msgid nntp.MessageID, group nntp.Newsgroup) {
	h.sendArticle(msgid, group)
}

func (h *httpWebhook) sendArticle(msgid nntp.MessageID, group nntp.Newsgroup) {
	f, err := h.storage.OpenArticle(msgid.String())
	if err == nil {
		u, _ := url.Parse(h.conf.URL)
		var r *http.Response
		var ctype string
		if h.conf.Dialect == "vichan" {
			c := textproto.NewConn(f)
			var hdr textproto.MIMEHeader
			hdr, err = c.ReadMIMEHeader()
			if err == nil {
				var body io.Reader
				ctype = hdr.Get("Content-Type")
				if ctype == "" || strings.HasPrefix(ctype, "text/plain") {
					ctype = "text/plain"
				}
				ctype = strings.Replace(strings.ToLower(ctype), "multipart/mixed", "multipart/form-data", 1)
				q := u.Query()
				for k, vs := range hdr {
					for _, v := range vs {
						q.Add(k, v)
					}
				}
				q.Set("Content-Type", ctype)
				u.RawQuery = q.Encode()

				if strings.HasPrefix(ctype, "multipart") {
					pr, pw := io.Pipe()
					log.Debug("using pipe")
					go func(in io.Reader, out io.WriteCloser) {
						_, params, _ := mime.ParseMediaType(ctype)
						if params == nil {
							// send as whatever lol
							io.Copy(out, in)
						} else {
							boundary, _ := params["boundary"]
							mpr := multipart.NewReader(in, boundary)
							mpw := multipart.NewWriter(out)
							mpw.SetBoundary(boundary)
							for {
								part, err := mpr.NextPart()
								if err == io.EOF {
									err = nil
									break
								} else if err == nil {
									// get part header
									h := part.Header
									// rewrite header part for php
									cd := h.Get("Content-Disposition")
									r := regexp.MustCompile(`filename="(.*)"`)
									// YOLO
									parts := r.FindStringSubmatch(cd)
									if len(parts) > 1 {
										fname := parts[1]
										h.Set("Content-Disposition", fmt.Sprintf(`filename="%s"; name="attachment[]"`, fname))
									}
									// make write part
									wp, err := mpw.CreatePart(h)
									if err == nil {
										// write part out
										io.Copy(wp, part)
									} else {
										log.Errorf("error writng webhook part: %s", err.Error())
									}
								}
								part.Close()
							}
							mpw.Close()
						}
						out.Close()
					}(c.R, pw)
					body = pr
				} else {
					body = f
				}
				r, err = http.Post(u.String(), ctype, body)
			}
		} else {
			var sz int64
			sz, err = f.Seek(0, 2)
			if err != nil {
				return
			}
			f.Seek(0, 0)
			// regular webhook
			ctype = "text/plain; charset=UTF-8"
			cl := new(http.Client)
			r, err = cl.Do(&http.Request{
				ContentLength: sz,
				URL:           u,
				Method:        "POST",
				Body:          f,
			})
		}
		if err == nil && r != nil {
			dec := json.NewDecoder(r.Body)
			result := make(map[string]interface{})
			err = dec.Decode(&result)
			if err == nil || err == io.EOF {
				msg, ok := result["error"]
				if ok {
					log.Warnf("hook gave error: %s", msg)
				} else {
					log.Debugf("hook response: %s", result)
				}
			} else {
				log.Warnf("hook response does not look like json: %s", err)
			}
			r.Body.Close()
			log.Infof("hook called for %s", msgid)
		}
	} else {
		f.Close()
	}
	if err != nil {
		log.Errorf("error calling web hook %s: %s", h.conf.Name, err.Error())
	}
}
