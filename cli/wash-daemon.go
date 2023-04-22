package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arwassa/wash-watch/pkg/washd"
	"github.com/spf13/viper"

	"golang.org/x/net/context"
	"golang.org/x/term"
	"google.golang.org/grpc"

	"github.com/spf13/pflag"
)

type FlagE struct {
	*flag.Flag
}

func (flag FlagE) HasChanged() bool {
	return false
}
func (flag FlagE) Name() string {
	return flag.Flag.Name
}
func (flag FlagE) ValueString() string {
	return flag.Value.String()
}
func (flag FlagE) ValueType() string {
	return "string"
}

func main() {

	applicationContext, cancelApplicationContext := context.WithCancel(context.Background())
	termSignal := make(chan os.Signal, 1)

	pflag.String("username", "", "Username")
	pflag.String("password", "", "Password")
	pflag.Bool("http", false, "listen on http")
	pflag.Parse()
	viper.SetEnvPrefix("WASHD")
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	viper.BindPFlag("username", pflag.Lookup("username"))
	viper.BindPFlag("password", pflag.Lookup("password"))
	viper.BindPFlag("http", pflag.Lookup("http"))
	viper.ReadInConfig()
	viper.AutomaticEnv()
	fmt.Printf("%s %s\n", viper.GetString("username"), viper.GetString("password"))
	if !viper.IsSet("username") {
		fmt.Print("Username: ")
		username, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(err)
		}
		viper.Set("username", username)
	}
	if !viper.IsSet("password") {
		fmt.Print("Password: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			panic(err)
		}
		viper.Set("password", string(password))
	}
	fmt.Println("Starting...")

	s := grpc.NewServer()
	washd.RegisterWashServiceServer(s, washd.CreateWashServiceServerHandle(washd.LoginCredentials{
		Username: viper.GetString("username"),
		Password: viper.GetString("password"),
	}))

	signal.Notify(termSignal, syscall.SIGINT, syscall.SIGTERM)
	var servErr error

	var listen net.Listener
	if viper.GetBool("http") {
		cert, err := tls.LoadX509KeyPair("washd-cert.pem", "washd-key.pem")
		if err != nil {
			panic(err)
		}
		listen, err = tls.Listen("tcp", "localhost:5551", &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2"},
			ServerName:   "localhost",
		})
		if err != nil {
			panic(err)
		}
	} else {
		var err error
		listen, err = net.Listen("unix", "./test_sock.sock")
		if err != nil {
			err = os.Remove("./test_sock.sock")
			if err != nil {
				panic(err)
			}
			listen, err = net.Listen("unix", "./test_sock.sock")
			if err != nil {
				panic(err)
			}
		}
	}
	defer listen.Close()

	go func() {
		if viper.GetBool("http") {
			servErr = http.Serve(listen, s)
		} else {
			servErr = s.Serve(listen)
		}
		cancelApplicationContext()
	}()
	go func() {
		select {
		case <-termSignal:
			fmt.Println("Terminating...")
			cancelApplicationContext()
		case <-applicationContext.Done():
		}
	}()

	<-applicationContext.Done()
	s.GracefulStop()
	if applicationContext.Err() != nil {
		fmt.Println(applicationContext.Err())
	}
	if servErr != nil {
		fmt.Println(servErr)
	}
}
