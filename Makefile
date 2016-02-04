dockerBuild:
	docker build -t goprox .

dockerRun:
	docker run -itdP \
      		-p 80:80 \
        	-e "PORT=80" \
        	goprox	
