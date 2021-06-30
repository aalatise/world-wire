FROM node:10-stretch AS builder
USER node
WORKDIR /home/node
COPY --chown=node:node package.json package-lock.json angular.json ./
RUN npm install
COPY --chown=node:node . .
RUN npm run build
ENV TZ=Europe/London
EXPOSE 4200
CMD npm run startInDocker
