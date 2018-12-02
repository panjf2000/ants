// MIT License

// Copyright (c) 2018 Andy Pan

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package test_http

import (
	"io/ioutil"
	"net/http"

	"github.com/panjf2000/ants"
)

type Request struct {
	Param  []byte
	Result chan []byte
}

func main() {
	pool, _ := ants.NewPoolWithFunc(100, func(payload interface{}) {
		request, ok := payload.(Request)
		if !ok {
			request = Request{Param:[]byte(""), Result: make(chan []byte)}
		}
		reverseParam := func(s []byte) []byte {
			for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
				s[i], s[j] = s[j], s[i]
			}
			return s
		}(request.Param)

		request.Result <- reverseParam
	})
	defer pool.Release()

	http.HandleFunc("/reverse", func(w http.ResponseWriter, r *http.Request) {
		param, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "request error", http.StatusInternalServerError)
		}
		defer r.Body.Close()

		request := Request{Param: param, Result: make(chan []byte)}

		// Throttle the requests with ants pool. This process is asynchronous and
		// you can receive a result from the channel defined outside.
		if err := pool.Serve(request); err != nil {
			http.Error(w, "throttle limit error", http.StatusInternalServerError)
		}

		w.Write(<-request.Result)
	})

	http.ListenAndServe(":8080", nil)
}
