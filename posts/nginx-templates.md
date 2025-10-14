---
title: "Nginx Templates and Examples"
date: "2025-07-30"
---

# Nginx Templates and Examples

This repo is a collection of templates and examples for how to use Nginx as a web server, load-balancer, and reverse proxy. Most of these ideas and concepts were learned from [this great tutorial](https://www.youtube.com/watch?v=q8OleYuqntY&ab_channel=TechWorldwithNana).

## Templates

This section contains templates for the following Nginx use cases

- Web Server
- Reverse Proxy
- Load Balancer
- HTTPS Redirect
- Caching

The following templates show the structure of different Nginx configurations but are not actually working files.

### Basic Web Server Template

Nginx is a web server technology wher the main config file is named "nginx.conf" and is usually placed in the `/etc/nginx/` folder.

The following is a basic template for a web server. The file defines a server listening on port 80 and serves the index.html file for any route.

```sh
# nginx.conf
server {
	listen 80;
	server_name example.com www.example.com;

	location / {
		root /var/www/example.com;
		index index.html index.htm;
	}
}
```

### Basic Reverse Proxy Template

Nginx can be used as a reverse proxy to sit in front of your services and route traffic. The advantage of this is it reduces the attack surface of your application and allows you to focus on the security of a single entry point.

In this example we define a server and then define a backend service where we want all traffic to be forwarded to

```sh
# nginx.conf
server {
	listen 80
	server_name api.example.com;

	location / {
		proxy_pass http://backend_server_address/;
		proxy_set_header Host $host;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwared-Proto $scheme;
	}
}
```

NOTE: it is very important to have the trailing `/` on the proxy address!

### Load Balancer Template

Extending the idea of using Nginx as a reverse proxy, we can use it as a load balancer. We can run multiple replicas of our services and use Nginx to distribute traffic between them for better reliability. Nginx also allows you to specify different algorithms for balancing the load like choosing the server with the least connections.

```sh
# nginx.conf
http {
    # 3 instances of the same app running
    upstream myapp1 {
        # choose the load balancing algo - default: round_robin
        # least_conn;
        server srv1.example.com;
        server srv2.example.com;
        server srv3.example.com;
    }

    server {
        listen 80;

        location / {
            # proxy all requests to the myapp1 server group
            proxy_pass http://myapp1/;
        }
    }
}
```

### HTTPS Redirect Template

When it comes to security we want to serve our application on HTTPS. We also want to redirect HTTP requests to HTTPS for a good user experience. This HTTPS redirect template shows how to set this up.

Define one server to listen on port 80 (default HTTP port) and redirect traffic to HTTPS with a 301 code. The next server listens on port 443 (default HTTPS port), loads the SSL certificates and serves the application.

```sh
# nginx.conf
server {
    listen 80;
    server_name: example.com www.example.com;

    # Redirect HTTP requests to HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name: example.com www.example.com;

    # Configure SSL
    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem
    ssl_certificate_key /etc/letsencrypt/live/example.com/privatekey.pem

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; include_subdomains always;"

    location / {
        root /var/www/example.com;
        index index.html index.htm;
    }
}
```

### Caching Template

This template shows how to use the `proxy_cache_path` directive to define the path for storing the cache and naming it `mycache` and then storing data for 10 minutes.

```sh
# nginx.conf
http {
    # cache server responses to the zone "mycache"
    proxy_cache_path /data/nginx/cache keys_zone=mycache:10m;

    server {
        proxy_cache: mycache;
        location / {
            proxy_pass https://localhost:8000/;
        }
    }
}
```

## Examples

The following examples can be run locally by editing the `nginx.conf` on your local machine and then reloading the server with `nginx -s reload`. Check out the Go example to see how to run the Nginx proxy and Go server with Docker and Docker Compose.

You can find the location of your Nginx install with `nginx -V` and finding the `--conf-path`. For me it is `/opt/homebrew/etc/nginx/nginx.conf`.

This section contains examples for the following Nginx use cases

- Static Site Server Example
- Reverse Proxy for a Node Server with SSL Example
- Reverse Proxy with Multiple Services Example
- Load Balancer for a Vue Frontend and Node Backend Application
- Reverse Proxy for Go Server with SSL

### Self-Signed SSL Certificate

For examples using SSL you will also need to generate your own self-signed SSL certificate for access to the server over HTTPS using `openssl`

`openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout nginx-selfsigned.key -out nginx-selfsigned.crt`

x509 certificate type
nodes - do not encrypt the private key with a passphrase
days 365 - certificate is valid for a year
newkey rsa:2048 creates a 2048 bit RSA key pair

### Static Site Server

With this example we use the NGINX docker image and docker compose to host a static website. This idea can be used to host an SPA by simply putting the built distribution of your application in the expected directory for static hosting.

The NGINX base image looks something like the first server template above and is configured to serve files from `/usr/share/nginx/html` by default. We use docker compose to bind our folder containing an html file to the expected location in the container.

The base image also listens on port 80 by default. We can use docker compose to bind any port we like to send traffic port 80. I chose 8089 here.

```yml
# docker-compose.yaml
services:
  nginx:
    image: nginx:1-alpine
    ports:
      - 8089:80
    volumes:
      - ./static/:/usr/share/nginx/html
```

Test it out by running `docker compose up` from the `/docker-compose-static` directory and then going to `localhost:8089` in your browser. You should see a basic hello world page.

### Reverse Proxy for Go Server with SSL

[Source code](https://github.com/tmastrom/nginx-examples/tree/main/go-example)

Create a basic Go server

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Go in Docker!")
	})

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		for name, headers := range r.Header {
			for _, h := range headers {
				fmt.Fprintf(w, "%v: %v\n", name, h)
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})

	fmt.Println("Server starting on port 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

```

Create a `nginx.conf` file that looks to pass traffic to the go server

```sh
worker_processes 1;

events {
    worker_connections 1024;
}

http {
    include mime.types;

    server {
        listen 443 ssl;

        ssl_certificate /etc/nginx/nginx-selfsigned.crt;
        ssl_certificate_key /etc/nginx/nginx-selfsigned.key;

        location / {
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            proxy_pass http://backend:8181/;
        }
    }

    # Redirect HTTP requests to HTTPS
    server {
        listen 80;

        location / {
            return 301 https://$host$request_uri;
        }
   }
}
```

Then create a docker compose file that runs the Nginx server in a container and the go server in another container. Make sure to mount the SSL certificates and ensure both containers are on the same network.

Run `docker compose up --build` and access the Go server APIs at `localhost:80` or `localhost:443`. Try out the different routes `/hello`, `/headers`, `/static`

Check out the code on [GitHub](https://github.com/tmastrom/nginx-examples/tree/main) for more examples.
