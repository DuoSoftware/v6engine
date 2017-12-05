#!/bin/bash

cd /home/DUO_V6ENGINE
if [ "$DEPLOYMENT_GROUP_NAME" == "staging" ]; then
    cp scripts/staging_agent.config /home/DUO_V6_OBSTORE/agent.config
    cp scripts/staging_agent.config /home/DUO_V6_AUTH/agent.config
    cp scripts/staging_agent.config /home/DUO_V6_NOTIFIER/agent.config
    cp scripts/staging_auth.config /home/DUO_V6_AUTH/Auth.config
     
    cp scripts/Terminal.config /home/DUO_V6_AUTH/Terminal.config
    cp scripts/settings.config /home/DUO_V6_AUTH/settings.config
    cp scripts/settings.config /home/DUO_V6_NOTIFIER/settings.config
    cp scripts/auth_dockerfile /home/DUO_V6_AUTH/Dockerfile
    cp scripts/notifier_dockerfile /home/DUO_V6_NOTIFIER/Dockerfile
    cp scripts/obstore_dockerfile /home/DUO_V6_OBSTORE/Dockerfile
    
    cp ObjectStore /home/DUO_V6_OBSTORE/
    cp duoauth/duoauth /home/DUO_V6_AUTH/
    cp DuoNotifier /home/DUO_V6_NOTIFIER/
        
elif [ "$DEPLOYMENT_GROUP_NAME" == "production" ]; then
    cp scripts/production_agent.config /home/DUO_V6_OBSTORE/agent.config
    cp scripts/production_agent.config /home/DUO_V6_AUTH/agent.config
    cp scripts/production_agent.config /home/DUO_V6_NOTIFIER/agent.config
    cp scripts/production_auth.config /home/DUO_V6_AUTH/Auth.config
    
    cp scripts/auth_dockerfile /home/DUO_V6_AUTH/Dockerfile
    cp scripts/notifier_dockerfile /home/DUO_V6_NOTIFIER/Dockerfile
    cp scripts/obstore_dockerfile /home/DUO_V6_OBSTORE/Dockerfile
     
    cp scripts/Terminal.config /home/DUO_V6_AUTH/Terminal.config
    cp scripts/settings.config /home/DUO_V6_AUTH/settings.config
    cp scripts/settings.config /home/DUO_V6_NOTIFIER/settings.config
    
    cp ObjectStore /home/DUO_V6_OBSTORE/
    cp duoauth/duoauth /home/DUO_V6_AUTH/
    cp DuoNotifier /home/DUO_V6_NOTIFIER/
fi
