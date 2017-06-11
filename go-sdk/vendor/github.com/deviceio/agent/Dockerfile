FROM centos:7

RUN yum -y install wget

RUN wget https://github.com/deviceio/agent/releases/download/test.v1/deviceio-agent.linux.amd64 \
        -O ~/agent && \
        chmod +x ~/agent

RUN printf '#!/bin/bash\n\
if [ -z "$HUB_HOST" ]; then \n\
        HUB_HOST="127.0.0.1" \n\
fi \n\
if [ -z "$HUB_PORT" ]; then \n\
        HUB_PORT=8975 \n\
fi \n\
~/agent install -o container -h $HUB_HOST -p $HUB_PORT -i \n\
/opt/deviceio/agent/container/bin/deviceio-agent start /opt/deviceio/agent/container/config.json \n\
'\
>> ~/run.sh

CMD sh ~/run.sh