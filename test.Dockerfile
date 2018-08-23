FROM debian:jessie

RUN apt-get -qqy update && \
	apt-get -qqy install jq openssl ca-certificates && \
        apt-get -qqy clean && \
        rm -rf /var/lib/apt/lists/*

# Create the directories and files that need to be present
RUN mkdir -p /var/lib/waagent && \
        mkdir -p /var/lib/waagent/Extension/config && \
        mkdir -p /var/lib/waagent/agent && \
        touch /var/lib/waagent/Extension/config/0.settings && \
        mkdir -p /var/lib/waagent/Extension/status && \
        mkdir -p /var/log/azure/Extension/VE.RS.ION

RUN chown root /var/lib/waagent && \
        chown root /var/lib/waagent/Extension/config && \
        chown root /var/lib/waagent/agent && \
        chown root /var/lib/waagent/Extension/config/0.settings && \
        chown root /var/lib/waagent/Extension/status && \
        chown root /var/log/azure/Extension/VE.RS.ION

# cd into the new workdir
WORKDIR /var/lib/waagent

# Copy the test environment
COPY integration-test/env/ .

RUN ln -s /var/lib/waagent/fake-waagent /sbin/fake-waagent && \
        ln -s /var/lib/waagent/wait-for-enable /sbin/wait-for-enable

# Copy the handler files
COPY misc/HandlerManifest.json ./Extension/
COPY misc/guest-configuration-shim ./Extension/bin/
COPY bin/guest-configuration-extension ./Extension/bin/
COPY integration-test/testdata/DesiredStateConfiguration-test.zip ./agent/DesiredStateConfiguration_1.0.0.zip

RUN chown root ./Extension/bin/
RUN chmod 777 ./Extension/bin/*