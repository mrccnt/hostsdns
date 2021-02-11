package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"io/ioutil"
	"log"
	"net"
	"os"
)

// Config is the hostsdns configuration file
type Config struct {
	Bind    string            `json:"bind"`
	Port    int               `json:"port"`
	DNS     string            `json:"dns"`
	Records map[string]string `json:"records"`
}

// cfg is our globaly available *Config
var cfg *Config

func main() {

	log.Println("Loading configs")
	if err := config(); err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("Starting up @ %s:%d\n", cfg.Bind, cfg.Port)
	server, err := net.ListenUDP("udp", &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP(cfg.Bind)})
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Forwarding host is", cfg.DNS)

	for {

		bytes := make([]byte, 2048)

		n, remoteaddr, err := server.ReadFromUDP(bytes)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		go handle(server, remoteaddr, bytes[:n])
	}
}

func handle(conn *net.UDPConn, addr *net.UDPAddr, bytes []byte) {

	packet := gopacket.NewPacket(bytes, layers.LayerTypeDNS, gopacket.Default)
	dnsPacket := packet.Layer(layers.LayerTypeDNS)
	layer, isLayer := dnsPacket.(*layers.DNS)
	if !isLayer {
		log.Println("Dropping non DNS query")
		return
	}

	reqname := string(layer.Questions[0].Name)

	if ip, ok := cfg.Records[reqname]; ok {
		if layer.Questions[0].Type == layers.DNSTypeA || layer.Questions[0].Type == layers.DNSTypeAAAA {
			log.Println("Self-Answering", layer.Questions[0].Type.String(), "request for", reqname)
			answer(conn, addr, layer, ip)
			return
		}
	}

	log.Println("Forwarding", layer.Questions[0].Type.String(), "request for", reqname)
	forward(conn, addr, bytes)
}

func forward(conn *net.UDPConn, addr *net.UDPAddr, bytes []byte) {

	client, err := net.Dial("udp", cfg.DNS)
	if err != nil {
		log.Println("[ERROR] [Forwarding] [netDial]", err.Error())
		return
	}
	defer func() { _ = client.Close() }()

	if _, err = fmt.Fprintf(client, string(bytes)); err != nil {
		log.Println("[ERROR] [Forwarding] [fmt.Fprintf]", err.Error())
		return
	}

	p := make([]byte, 2048)

	n, err := bufio.NewReader(client).Read(p)
	if err != nil {
		log.Println("[ERROR] [Forwarding] [Read]", err.Error())
		return
	}

	//dumpbytes("layer.forward.json", p[:n])

	if _, err = conn.WriteToUDP(p[:n], addr); err != nil {
		log.Println("[ERROR] [Forwarding] [WriteToUDP]", err.Error())
	}
}

func answer(conn *net.UDPConn, addr *net.UDPAddr, layer *layers.DNS, ip string) {

	parsedCIDR, _, err := net.ParseCIDR(ip + "/24")
	if err != nil {
		log.Println("[ERROR] [handle] [net.ParseCIDR]", err.Error())
		return
	}

	layer.ResponseCode = layers.DNSResponseCodeNoErr
	layer.Answers = nil
	if layer.Questions[0].Type == layers.DNSTypeA {
		layer.Answers = []layers.DNSResourceRecord{
			{
				Name:       layer.Questions[0].Name,
				Type:       layer.Questions[0].Type,
				Class:      layer.Questions[0].Class,
				IP:         parsedCIDR,
				Data:       []byte{}, // TODO
				DataLength: 0,        // TODO
			},
		}
	}
	layer.QR = true
	layer.ANCount = uint16(len(layer.Answers))
	layer.OpCode = layers.DNSOpCodeQuery
	layer.RA = true

	// dumplayer("layer.anser.json", layer)

	buf := gopacket.NewSerializeBuffer()
	if err = layer.SerializeTo(buf, gopacket.SerializeOptions{}); err != nil {
		log.Println("[ERROR] [handle] [SerializeTo]", err.Error())
		return
	}
	if _, err = conn.WriteTo(buf.Bytes(), addr); err != nil {
		log.Println("[ERROR] [handle] [WriteTo]", err.Error())
	}

}

func config() error {

	var f string
	var bytes []byte
	var err error

	flag.StringVar(&f, "f", "", "The config.json file")
	flag.Parse()

	if _, err = os.Stat(f); err != nil {
		return err
	}

	if bytes, err = ioutil.ReadFile(f); err != nil {
		return err
	}

	cfg = new(Config)

	if err = json.Unmarshal(bytes, cfg); err != nil {
		return err
	}

	return nil
}

//func dumpbytes(path string, bytes []byte) {
//
//	packet := gopacket.NewPacket(bytes, layers.LayerTypeDNS, gopacket.Default)
//	dnsPacket := packet.Layer(layers.LayerTypeDNS)
//	layer, isLayer := dnsPacket.(*layers.DNS)
//	if !isLayer {
//		log.Println("[ERROR] [dumpbytes] [isLayer]", "non dns query")
//		return
//	}
//
//	dumplayer(path, layer)
//}
//
//func dumplayer(path string, layer *layers.DNS) {
//
//	_ = os.Remove(path)
//
//	bs, err := json.MarshalIndent(layer, "", " ")
//	if err != nil {
//		log.Println("[ERROR] [dumplayer] [MarshalIndent]", err.Error())
//		return
//	}
//
//	if err = ioutil.WriteFile(path, bs, 0644); err != nil {
//		log.Println("[ERROR] [dumplayer] [WriteFile]", err.Error())
//		return
//	}
//}
