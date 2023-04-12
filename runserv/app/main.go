package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	port    = flag.Int("port", 8080, "app server port")
	podName = flag.String("pod_name", "", "pod name")
	podIP   = flag.String("pod_ip", "", "pod IP")
	ctx     = context.Background()
)

var (
	router *gin.Engine
)

func init() {
	flag.Parse()

	router = gin.Default()
}

func HostIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("net.InterfaceAddr failed, err %v", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("failed to get container IP")
}

func serve() {
	rg := router.Group("")
	rg.GET("/echo", func(c *gin.Context) {
		config, err := rest.InClusterConfig()
		if err != nil {
			c.Abort()
			c.JSON(http.StatusOK, map[string]interface{}{
				"err": err,
			})
			return
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusOK, map[string]interface{}{
				"err": err,
			})
			return
		}

		pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			c.Abort()
			c.JSON(http.StatusOK, map[string]interface{}{
				"err": err,
			})
			return
		}

		sz := len(pods.Items)
		podInfos := make([]string, sz)

		for i, pod := range pods.Items {
			podInfos[i] = pod.Name + " - " + pod.Status.PodIP
		}

		hostIP, err := HostIP()
		c.JSON(http.StatusOK, map[string]interface{}{
			"port":     port,
			"podInfos": podInfos,
			"podName":  podName,
			"podIP":    podIP,
			"hostIP":   hostIP,
		})
	})

	router.Run(fmt.Sprintf(":%v", *port))
}

func main() {
	serve()
}
