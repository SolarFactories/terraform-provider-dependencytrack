package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	serverAddress string = "127.0.0.1:8079"
)

func main() {
	tlsConfig, err := getTLSConfig()
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()
	router.HandleFunc("/.well-known/openid-configuration", handleOpenIDConfiguration)
	router.HandleFunc("/.well-known/jwks", handleJwks)

	srv := http.Server{
		TLSConfig:    tlsConfig,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		Handler:      router,
		Addr:         serverAddress,
	}
	fmt.Println("Listening on", srv.Addr)
	err = srv.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatal(err)
	}
}

func handleOpenIDConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]any{
		"issuer":   "https://" + serverAddress,
		"jwks_uri": "https://" + serverAddress + "/.well-known/jwks",
	})
	if err != nil {
		fmt.Println("Error in encoding openid configuration: ", err.Error())
		return
	}
}

func handleJwks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(false) // TODO: Encode JWKS.
	if err != nil {
		fmt.Println("Error in encoding JWKS: ", err.Error())
		return
	}
}

func getTLSConfig() (*tls.Config, error) {
	certificate, err := genCertificate()
	if err != nil {
		return nil, err
	}
	config := tls.Config{
		Certificates: []tls.Certificate{*certificate},
		MinVersion:   tls.VersionTLS13,
	}
	return &config, nil
}

func genCertificate() (*tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, errors.New("Unable to generate Keypair: " + err.Error())
	}

	template := x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"Terraform Provider DependencyTrack"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(0, 0, 7),
		DNSNames:  []string{"localhost"},
		IsCA:      true,
		KeyUsage:  x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, errors.New("Unable to create certificate: " + err.Error())
	}
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, errors.New("Unable to load TLS keypair: " + err.Error())
	}
	return &cert, nil
}
