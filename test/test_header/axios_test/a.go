package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

func main() {
	bodyBuf := bytes.NewBuffer(nil)
	bodyBuf.WriteString(`------WebKitFormBoundaryWdDAe6hxfa4nl2Ig
Content-Disposition: form-data; name="submit-name"
test
------WebKitFormBoundaryWdDAe6hxfa4nl2Ig
Content-Disposition: form-data; name="file1"; filename="out.png"
Content-Type: image/png
binary-data
------WebKitFormBoundaryWdDAe6hxfa4nl2Ig
Content-Disposition: form-data; name="file2"; filename="2.png"
Content-Type: image/png
binary-data-2
------WebKitFormBoundaryWdDAe6hxfa4nl2Ig--`)
	req := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": {`multipart/form-data; boundary=----WebKitFormBoundaryWdDAe6hxfa4nl2Ig`}},
		Body:   ioutil.NopCloser(bodyBuf),
	}
	getMultiPart3(req)
}

//通过r.ParseMultipartForm
func GetMultiPart1(r *http.Request) {
	/**
	  底层通过调用multipartReader.ReadForm来解析
	  如果文件大小超过maxMemory,则使用临时文件来存储multipart/form中文件数据
	*/
	r.ParseMultipartForm(128)
	fmt.Println("r.Form:", r.Form)
	fmt.Println("r.PostForm:", r.PostForm)
	fmt.Println("r.MultiPartForm:", r.MultipartForm)
	GetFormData(r.MultipartForm)
}

//通过MultipartReader
func GetMultiPart2(r *http.Request) {
	mr, err := r.MultipartReader()
	if err != nil {
		fmt.Println("r.MultipartReader() err,", err)
		return
	}
	form, _ := mr.ReadForm(128)
	GetFormData(form)
}

//字节解析multi-part
func getMultiPart3(r *http.Request) {
	mr, err := r.MultipartReader()
	if err != nil {
		fmt.Println("r.MultipartReader() err,", err)
		return
	}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("mr.NextPart() err,", err)
			break
		}
		fmt.Println("part header:", p.Header)
		formName := p.FormName()
		fileName := p.FileName()
		if formName != "" && fileName == "" {
			formValue, _ := ioutil.ReadAll(p)
			fmt.Printf("formName:%s,formValue:%s\n", formName, formValue)
		}
		if fileName != "" {
			fileData, _ := ioutil.ReadAll(p)
			fmt.Printf("fileName:%s,fileData:%s\n", fileName, fileData)
		}
		fmt.Println()
	}
}

func GetFormData(form *multipart.Form) {
	//获取 multi-part/form body中的form value
	for k, v := range form.Value {
		fmt.Println("value,k,v = ", k, ",", v)
	}
	fmt.Println()
	//获取 multi-part/form中的文件数据
	for _, v := range form.File {
		for i := 0; i < len(v); i++ {
			fmt.Println("file part ", i, "-->")
			fmt.Println("fileName :", v[i].Filename)
			fmt.Println("part-header:", v[i].Header)
			f, _ := v[i].Open()
			buf, _ := ioutil.ReadAll(f)
			fmt.Println("file-content", string(buf))
			fmt.Println()
		}
	}
}
