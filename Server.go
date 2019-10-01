package main
import ("fmt"
		"github.com/gorilla/mux"
    	"net/http"
  )
func main(){
  r := mux.NewRouter()
  r.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to my website!")
	})
  http.Handle("/",r)
  http.ListenAndServe(":8090",nil)
}
