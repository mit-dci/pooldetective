FROM node:10 as nodebuild
RUN mkdir /tmp/build
COPY package.json /tmp/build/package.json
COPY package-lock.json /tmp/build/package-lock.json
WORKDIR /tmp/build
RUN npm install
COPY src/ /tmp/build/src/
COPY public/ /tmp/build/public/
COPY yarn.lock /tmp/build/yarn.lock
RUN mkdir /tmp/output
RUN npm install && npm run build
RUN mv build /tmp/output/

FROM nginx:alpine
COPY --from=nodebuild /tmp/output/build /var/www/html

COPY nginx.conf /etc/nginx/conf.d/default.conf 
