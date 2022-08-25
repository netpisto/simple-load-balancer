package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}
type Server interface{
	addresse() string 
	isAlive() bool
	serve(rw http.ResponseWriter,req *http.Request) 
}
func (s *simpleServer) addresse() string{
	return s.addr
}
func (s *simpleServer) isAlive() bool{
	return true
}
func (s *simpleServer) serve(rw http.ResponseWriter,req *http.Request){
	s.proxy.ServeHTTP(rw,req)
}
func newLoadBalancer(port string ,servers []Server) *LoadBalancer{
	return &LoadBalancer{
		port: port,
		servers: servers,
		counter: 0,
	}
}
func newSimpleServer(addr string) *simpleServer{
	serverUrl,err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr: addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func (lb *LoadBalancer) getNextAvailableServer()Server{
	server := lb.servers[lb.counter%len(lb.servers)]
	
	for !server.isAlive(){
		lb.counter++
		server = lb.servers[lb.counter%len(lb.servers)]
	}
	lb.counter++
	return server
}
func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter,req *http.Request){
	server := lb.getNextAvailableServer()
	log.Println(lb.counter,"-",req.RemoteAddr+" to "+server.addresse())
	server.serve(rw,req)
}

func handleErr(err error){
	if err != nil{
		log.Fatal(err)
	}
	
}

type LoadBalancer struct {
	port string
	counter int
	servers []Server
}

func main(){
	hosts := []string{}
	for true {
		fmt.Print("enter addresse to forward to (private or public): ")
		var addr string
		fmt.Scanln(&addr)
		if addr == ""{
			break
		}
		hosts = append(hosts, addr)
	}
	servers := []Server{}
	for i:= 0;i<len(hosts);i++{
		servers = append(servers,newSimpleServer(hosts[i]) )
	}
	lb := newLoadBalancer("8000",servers)
	server := http.Server{
		Addr: "0.0.0.0:"+lb.port,
	}
	http.HandleFunc("/",func(w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w,r)
	})
	res,_ := http.Get("https://ip4.seeip.org/")
	iodata,_:= ioutil.ReadAll(res.Body)
	ip := string(iodata)
	log.Println("serving at "+ip+"|0.0.0.0|localhost"+":"+lb.port)
	server.ListenAndServe()
}
