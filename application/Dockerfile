# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
FROM gobuffalo/buffalo:v0.18.8 as builder

ARG ci_commit_id

RUN mkdir -p $GOPATH/src/wecarry
WORKDIR $GOPATH/src/wecarry

# Install the desired version of NodeJS
RUN npm cache clean -f && \
    npm install -g n && \
    n 16

COPY . .

RUN sh -c "echo $ci_commit_id > domain/commit.txt"
RUN cat domain/commit.txt

RUN buffalo build --static -o /bin/app

FROM alpine
RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates

WORKDIR /bin/

COPY --from=builder /bin/app .

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# Uncomment to run the migrations before running the binary:
CMD /bin/app migrate; /bin/app
#CMD exec /bin/app
