FROM node:24-alpine AS build

WORKDIR /src
ARG VITE_API_BASE_URL=http://localhost:8080
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV CI=true

COPY package.json pnpm-lock.yaml ./
RUN corepack enable && corepack prepare pnpm@10.33.0 --activate && pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

FROM nginx:1.29-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html

EXPOSE 80
