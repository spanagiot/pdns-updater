package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"syscall"
	"time"
)

var host string

// IPAPIResponse describes the JSON response of my IP API
type IPAPIResponse struct {
	IP string
}

// IPStack describes an IP stack consisting of IPv4 and IPv6
type IPStack struct {
	IPv4 string
	IPv6 string
}

// RecordToUpdate contains the details of the record that will be updated
type RecordToUpdate struct {
	id       string
	host     string
	value    string
	password string
}

// GetExternalIPStack fetches the api
// and returns the external IP
func GetExternalIPStack() IPStack {
	var myExternalIP IPStack

	myIPv4, err := getExternalIPByNetworkStack("tcp4")
	if err != nil {
		myExternalIP.IPv4 = ""
	} else {
		myExternalIP.IPv4 = myIPv4
	}
	myIPv6, err := getExternalIPByNetworkStack("tcp6")
	if err != nil {
		myExternalIP.IPv6 = ""
	} else {
		myExternalIP.IPv6 = myIPv6
	}

	return myExternalIP
}

// getExternalIPByNetworkStack performs an HTTP request to the api using only
// the network stack we asked so we can get on demand the IPv4 or the IPv6
// Network stack can have the values "tcp4" or "tcp6"
func getExternalIPByNetworkStack(networkStack string) (string, error) {
	var MyTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
			Control: func(network, address string, c syscall.RawConn) error {
				if network != networkStack {
					return errors.New("Network used different than requested")
				}
				return nil
			},
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	var myClient = http.Client{Transport: MyTransport}
	resp, err := myClient.Get(fmt.Sprintf("https://%s/api/v1/remote/ip", host))
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var decodedResponse IPAPIResponse
	err = json.Unmarshal(body, &decodedResponse)
	if json.Valid(body) == false {
		fmt.Println("Invalid json body")
	}
	if err != nil {
		panic(err)
	}
	return decodedResponse.IP, nil
}

// GetInternalIPStack returns a struct with the
// IPv4 and the IPv6 of the machine on the
// given interface
func GetInternalIPStack(interfacePtr *string) (IPStack, error) {
	var internalIPs IPStack

	inter, err := net.InterfaceByName(*interfacePtr)
	if err != nil {
		switch err.Error() {
		case "route ip+net: no such network interface":
			return IPStack{}, errors.New("Interface not found")
		default:
			panic(err)
		}
	}
	addresses, err := inter.Addrs()
	if err != nil {
		panic(err)
	}
	for index := 0; index < len(addresses); index++ {
		switch v := addresses[index].(type) {
		case *net.IPNet:
			if !v.IP.IsLoopback() {
				if v.IP.To4() != nil {
					internalIPs.IPv4 = fmt.Sprintf("%s", v.IP.To4())
				} else if !v.IP.IsLinkLocalUnicast() {
					if v.IP.To16() != nil {
						internalIPs.IPv6 = fmt.Sprintf("%s", v.IP.To16())
					}
				}
			}
		}
	}

	return internalIPs, nil
}

func updateRecord(record RecordToUpdate) int {
	var url = fmt.Sprintf("https://%s/api/v1/remote/updatepw?record=%s&password=%s&content=%s", record.host, record.id, record.password, record.value)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func main() {
	iaIDPtr := flag.String("ia", "", "the ID of the internal A record to update")
	iaaaaIDPtr := flag.String("iaaaa", "", "the ID of the internal AAAA record to update")
	eaIDPtr := flag.String("ea", "", "the ID of the external A record to update")
	eaaaaIDPtr := flag.String("eaaaa", "", "the ID of the external AAAA record to update")

	passwrdPtr := flag.String("pass", "123456", "the password for the record")
	hostPtr := flag.String("host", "", "the host that implements the API")

	interfacePtr := flag.String("i", "en0", "the interface to get the internal IP")

	flag.Parse()

	iaID := string(*iaIDPtr)
	iaaaaID := string(*iaaaaIDPtr)
	eaID := string(*eaIDPtr)
	eaaaaID := string(*eaaaaIDPtr)

	password := string(*passwrdPtr)
	host = string(*hostPtr)

	// Check if we have a host to update
	if host == "" {
		fmt.Println("No host to be updated, my job is done.\nExiting")
		return
	}

	// Check if we need to update the internal IPs
	if iaID != "" || iaaaaID != "" {
		internalIPs, err := GetInternalIPStack(interfacePtr)
		if err != nil {
			fmt.Println(err)
			return
		}

		if iaID != "" {
			if internalIPs.IPv4 == "" {
				fmt.Printf("[INFO]No IPv4 found on the machine\n")
				internalIPs.IPv4 = "0.0.0.0"
			}
			var record = RecordToUpdate{id: iaID, host: host, value: internalIPs.IPv4, password: password}
			var statusCode = updateRecord(record)
			if statusCode < 300 {
				fmt.Printf("[UPDATED]host:%s,id:%s,type:A,value:%s\n", record.host, record.id, record.value)
			}

		}
		if iaaaaID != "" {
			if internalIPs.IPv6 == "" {
				fmt.Printf("[INFO]No IPv6 found on the machine\n")
				internalIPs.IPv6 = "::"
			}
			var record = RecordToUpdate{id: iaaaaID, host: host, value: internalIPs.IPv6, password: password}
			var statusCode = updateRecord(record)
			if statusCode < 300 {
				fmt.Printf("[UPDATED]host:%s,id:%s,type:AAAA,value:%s\n", record.host, record.id, record.value)
			}

		}
	}

	// Check if we need to update the external IPs
	if eaID != "" || eaaaaID != "" {
		var externalIP = GetExternalIPStack()

		if eaID != "" {
			if externalIP.IPv4 == "" {
				fmt.Printf("[INFO]No external IPv4 found on the machine\n")
				externalIP.IPv4 = "0.0.0.0"
			}
			var record = RecordToUpdate{id: eaID, host: host, value: fmt.Sprintf("%s", externalIP.IPv4), password: password}
			var statusCode = updateRecord(record)
			if statusCode < 300 {
				fmt.Printf("[UPDATED]host:%s,id:%s,type:A,value:%s\n", record.host, record.id, record.value)
			} else {
				switch statusCode {
				case 403:
					fmt.Printf("[ERROR]host:%s,id:%s,type:A,error:permissions\n", record.host, record.id)
				case 404:
					fmt.Printf("[ERROR]host:%s,id:%s,type:A,error:notfound\n", record.host, record.id)
				case 422:
					fmt.Printf("[ERROR]host:%s,id:%s,type:A,error:argumentmissing\n", record.host, record.id)

				}
			}
		}
		if eaaaaID != "" {
			if externalIP.IPv6 == "" {
				fmt.Printf("[INFO]No external IPv6 found on the machine\n")
				externalIP.IPv6 = "::"
			}
			var record = RecordToUpdate{id: eaaaaID, host: host, value: fmt.Sprintf("%s", externalIP.IPv6), password: password}
			var statusCode = updateRecord(record)
			if statusCode < 300 {
				fmt.Printf("[UPDATED]host:%s,id:%s,type:AAAA,value:%s\n", record.host, record.id, record.value)
			} else {
				switch statusCode {
				case 403:
					fmt.Printf("[ERROR]host:%s,id:%s,type:A,error:permissions\n", record.host, record.id)
				case 404:
					fmt.Printf("[ERROR]host:%s,id:%s,type:A,error:notfound\n", record.host, record.id)
				case 422:
					fmt.Printf("[ERROR]host:%s,id:%s,type:A,error:argumentmissing\n", record.host, record.id)

				}
			}
		}
	}

}
