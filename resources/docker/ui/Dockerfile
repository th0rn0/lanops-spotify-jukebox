FROM node:21-alpine3.18
LABEL maintainer="Thornton Phillis (dev@th0rn0.co.uk)"

WORKDIR /frontend

COPY ui/ .

RUN npm install

RUN npm run build

EXPOSE 3000

CMD [ "node", "/frontend/.output/server/index.mjs" ]