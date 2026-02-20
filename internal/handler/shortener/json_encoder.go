package shortener

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/types"
)

type RequestData struct {
	URL string `json:"url"`
}

type ResponseData struct {
	Result string `json:"result"`
}

type JSONEncoder interface {
	CreateShortURL(resp http.ResponseWriter, req *http.Request)
}

func (s Service) CreateShortURL(resp http.ResponseWriter, req *http.Request) {
	var (
		data     RequestData
		respData ResponseData
		rBody    bytes.Buffer
		err      error
		buf      *bytes.Buffer
	)
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	//Check request header
	compressed := req.Header.Get("Content-Encoding")
	switch compressed {
	case "gzip":
		if rBody, err = bodyDecompressGzip(req); err != nil {
			log.Fatalln(err)
		}
		if _, err = io.Copy(buf, &rBody); err != nil {
			log.Fatalln(err)
		}
	case "deflate":
		if rBody, err = bodyDecompressDeflate(req); err != nil {
			log.Fatalln(err)
		}
		if _, err = io.Copy(buf, &rBody); err != nil {
			log.Fatalln(err)
		}
	default:
		if _, err = io.Copy(buf, req.Body); err != nil {
			log.Fatalln(err)
		}
	}

	if err := json.NewDecoder(buf).Decode(&data); err != nil {
		log.Println("cannot decode request: ", err.Error())
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	id := Hashing([]byte(data.URL))
	respData.Result = string(s.ResultAddr) + "/" + id
	s.Stor.SetData(types.ShortURL(id), types.RawURL(data.URL))
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(resp).Encode(respData); err != nil {
		log.Println("cannot encode response: ", err.Error())
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
}

func bodyDecompressGzip(r *http.Request) (bytes.Buffer, error) {
	var buf bytes.Buffer

	rd, err := gzip.NewReader(r.Body)
	if err != nil {
		return bytes.Buffer{}, err
	}
	defer rd.Close()

	if _, err := buf.ReadFrom(rd); err != nil {
		return bytes.Buffer{}, err
	}
	return buf, nil
}

func bodyDecompressDeflate(r *http.Request) (bytes.Buffer, error) {
	var buf bytes.Buffer

	rd, err := zlib.NewReader(r.Body)
	if err != nil {
		return bytes.Buffer{}, err
	}
	defer rd.Close()

	if _, err := buf.ReadFrom(rd); err != nil {
		return bytes.Buffer{}, err
	}
	return buf, nil
}
