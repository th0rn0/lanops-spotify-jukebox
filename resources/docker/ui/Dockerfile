FROM node

WORKDIR /app

COPY ui/ .

RUN npm install

RUN npm run build

EXPOSE 3000

CMD [ "node", "/app/.output/server/index.mjs" ]