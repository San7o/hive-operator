#!/bin/sh
sudo docker rmi manager localhost:5001/manager:latest &2>/dev/null || :
sudo docker build -t manager .
sudo docker tag manager localhost:5001/manager:latest
sudo docker push localhost:5001/manager:latest
