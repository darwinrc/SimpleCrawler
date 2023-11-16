FROM node
WORKDIR /app
COPY . .

RUN npm install

EXPOSE 3000
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
