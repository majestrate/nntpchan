#TinyHtml

A small html minimizer for use in web servers. Generally reduces size of code around 10-15%

Usage:

	file,_ := os.Open("Somefile.html")
	minHtml := tinyhtml.New(file)

	//Assuming you are using it in a http request handler
	http.ServeContent(w,r,"My Page", time.Now(), minHtml)

The Minimizer class is a wrapper over any io.Reader interface that also implements the io.Reader interface.
