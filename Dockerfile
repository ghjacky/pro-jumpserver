FROM alpine
MAINTAINER MyGuo <myguo@aibee.com>
ARG packagefile
ARG exposeport
ARG appname
ENV SERVER_BIN=/opt/${appname}/$appname \
	CONFIG_FILE=/opt/${appname}/configs/config.toml
COPY Shanghai /etc/localtime
RUN mkdir -p /opt/${appname} \
	echo "Asia/Shanghai" >  /etc/timezone
ADD ${packagefile} /opt/${appname}
CMD ["sh", "-c", "$SERVER_BIN --config $CONFIG_FILE"]
EXPOSE ${exposeport}