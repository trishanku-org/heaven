/*
Copyright 2023 Amshuman K R <amshuman.kr@gmail.com>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

type Certificate struct {
	PrivateKey     *rsa.PrivateKey
	PrivateKeyPEM  []byte
	Certificate    *x509.Certificate
	CertificatePEM []byte
	CA             *Certificate
	IsCA           bool
}

type CertificateConfig struct {
	CommonName         string
	Organization       []string
	OrganizationalUnit []string
	DNSNames           []string
	IsCA               bool
	CA                 *Certificate
}

const (
	EncryptionKeySize = 2048
	Expiration        = 365 * 24 * time.Hour

	BlockTypeCertificate   = "CERTIFICATE"
	BlockTypeRSAPrimaryKey = "RSA PRIVATE KEY"

	KeyCertificate   = corev1.TLSCertKey
	KeyPrimaryKey    = corev1.TLSPrivateKeyKey
	KeyCACertificate = corev1.ServiceAccountRootCAKey
	KeyCAPrimaryKey  = "ca.key"
	KeyKubeconfig    = "kubeconfig"
)

func (cc *CertificateConfig) Generate() (*Certificate, error) {
	var (
		c = &Certificate{
			CA:   cc.CA,
			IsCA: cc.IsCA,
		}

		now = time.Now()

		err            error
		slNo           *big.Int
		signerCert     *Certificate
		certificateDER []byte
	)

	if c.PrivateKey, err = rsa.GenerateKey(rand.Reader, EncryptionKeySize); err != nil {
		return nil, err
	}

	if slNo, err = rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128)); err != nil {
		return nil, err
	}

	c.Certificate = &x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  cc.IsCA,
		SerialNumber:          slNo,
		NotBefore:             now,
		NotAfter:              now.Add(Expiration),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		Subject: pkix.Name{
			CommonName:         cc.CommonName,
			Organization:       cc.Organization,
			OrganizationalUnit: cc.OrganizationalUnit,
		},
		DNSNames: cc.DNSNames,
	}

	if cc.IsCA {
		c.Certificate.KeyUsage |= x509.KeyUsageCRLSign | x509.KeyUsageCertSign
	} else {
		c.Certificate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	}

	if cc.CA == nil { // Self-signed.
		signerCert = c
	} else {
		signerCert = cc.CA
	}

	if certificateDER, err = x509.CreateCertificate(
		rand.Reader,
		c.Certificate,
		signerCert.Certificate,
		&c.PrivateKey.PublicKey,
		signerCert.PrivateKey,
	); err != nil {
		return nil, err
	}

	c.CertificatePEM = pem.EncodeToMemory(&pem.Block{
		Type:  BlockTypeCertificate,
		Bytes: certificateDER,
	})

	c.PrivateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  BlockTypeRSAPrimaryKey,
		Bytes: x509.MarshalPKCS1PrivateKey(c.PrivateKey),
	})

	return c, err
}

func (c *Certificate) WriteToMap(m map[string][]byte) {
	if c.IsCA {
		m[KeyCACertificate] = c.CertificatePEM
		m[KeyCAPrimaryKey] = c.PrivateKeyPEM
	} else {
		m[KeyCertificate] = c.CertificatePEM
		m[KeyPrimaryKey] = c.PrivateKeyPEM
		if c.CA != nil {
			m[KeyCACertificate] = c.CA.CertificatePEM
		}
	}
}

func LoadCertificate(primaryKeyPEM, certificatePEM []byte) (*Certificate, error) {
	var (
		c = &Certificate{
			PrivateKeyPEM:  primaryKeyPEM,
			CertificatePEM: certificatePEM,
		}

		block *pem.Block
		err   error
	)

	if block, _ = pem.Decode(primaryKeyPEM); block == nil || block.Type != BlockTypeRSAPrimaryKey {
		return nil, errors.New("error decoding RSA private key")
	}

	if c.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		return nil, err
	}

	if block, _ = pem.Decode(certificatePEM); block == nil || block.Type != BlockTypeCertificate {
		return nil, errors.New("error decoding certificate")
	}

	if c.Certificate, err = x509.ParseCertificate(block.Bytes); err != nil {
		return nil, err
	}

	return c, err
}

func LoadCertificateFromMap(m map[string][]byte) (*Certificate, error) {
	var (
		primaryKeyPEM, certificatePEM []byte
		ok                            bool
	)

	if primaryKeyPEM, ok = m[KeyPrimaryKey]; !ok {
		primaryKeyPEM = m[KeyCAPrimaryKey]
	}

	if certificatePEM, ok = m[KeyCertificate]; !ok {
		certificatePEM = m[KeyCACertificate]
	}

	return LoadCertificate(primaryKeyPEM, certificatePEM)
}

func GenerateKubeconfigAndWriteToMap(
	contextName, clusterName, userName string,
	cluster clientcmdv1.Cluster,
	authInfo clientcmdv1.AuthInfo,
	m map[string][]byte,
) (err error) {
	var kubeconfigRaw []byte

	if kubeconfigRaw, err = runtime.Encode(clientcmdlatest.Codec, &clientcmdv1.Config{
		CurrentContext: contextName,
		Clusters: []clientcmdv1.NamedCluster{{
			Name:    clusterName,
			Cluster: cluster,
		}},
		AuthInfos: []clientcmdv1.NamedAuthInfo{{
			Name:     userName,
			AuthInfo: authInfo,
		}},
		Contexts: []clientcmdv1.NamedContext{{
			Name: contextName,
			Context: clientcmdv1.Context{
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		}},
	}); err != nil {
		return
	}

	m[KeyKubeconfig] = kubeconfigRaw

	return
}
