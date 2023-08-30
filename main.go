package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

func createNewNginxFile(domainName string) {
	templateString := `
server {
	listen 80;
	listen [::]:80;

	gzip on;
	gzip_static on;

	index index.html index.htm index.nginx-debian.html;

	server_name {{ .ServerName }};

	location / {
			proxy_pass http://localhost:3000;
			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection 'upgrade';
			proxy_set_header Host $host;
			proxy_cache_bypass $http_upgrade;
	}

	location /api {
			proxy_pass http://localhost:8080;
			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection '';
			proxy_set_header Host $host;
			proxy_cache_bypass $http_upgrade;
			chunked_transfer_encoding off;
	}
}`

	templateString = strings.TrimLeft(templateString, "\n")

	fmt.Println("=> creating template nginx...")
	tmpl, err := template.New("nginxConfig").Parse(templateString)
	if err != nil {
		panic(err)
	}

	data := struct {
		ServerName string
	}{
		ServerName: domainName,
	}

	// currentDir, _ := os.Getwd() // for local testing

	fmt.Println("=> saving template nginx to target dir...")
	configFile := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domainName)
	// configFile := fmt.Sprintf("%s/%s.conf", currentDir, domainName) // for local testing

	file, err := os.Create(configFile)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		panic(err)
	}

	fmt.Println("=> create symbolink template nginx to target dir...")
	enabledFile := fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", domainName)
	// enabledFile := fmt.Sprintf("%s/enable/%s.conf", currentDir, domainName) // for local testing

	err = os.Symlink(configFile, enabledFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("=> testing nginx configuration...")
	nginxTestCmd := "nginx -t"
	err = execCommand(nginxTestCmd)
	if err != nil {
		panic(err)
	}

	fmt.Println("=> restart nginx service...")
	nginxRestartCmd := "sudo systemctl restart nginx"
	err = execCommand(nginxRestartCmd)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	fmt.Println("=> create certificate letsencrypt...")
	certbotCmd := fmt.Sprintf("sudo certbot --nginx -n --cert-name %s -d %s", domainName, domainName)
	err = execCommand(certbotCmd)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	fmt.Println("=> reload nginx service...")
	nginxReloadCmd := "sudo systemctl reload nginx"
	err = execCommand(nginxReloadCmd)
	if err != nil {
		panic(err)
	}

	fmt.Println("=> finished...")
}

func execCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Please enter a custom domain name")
		return
	}

	customDomain := os.Args[1]
	createNewNginxFile(customDomain)
}
