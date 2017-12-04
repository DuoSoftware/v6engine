#!/bin/bash

cd /home/DUO_V6ENGINE
if [ "$DEPLOYMENT_GROUP_NAME" == "staging" ]; then
    cp scripts/staging_agent.config /home/DUO_V6_OBSTORE/agent.config
    cp scripts/staging_agent.config /home/DUO_V6_AUTH/agent.config
    cp ObjectStore /home/DUO_V6_OBSTORE/
    cp duoauth/duoauth /home/DUO_V6_AUTH
    
    
elif ["$DEPLOYMENT_GROUP_NAME" == "production" ]; then
   cp scripts/production_agent.config /home/DUO_V6_OBSTORE/agent.config
fi
