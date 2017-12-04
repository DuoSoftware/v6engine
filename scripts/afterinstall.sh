cd /home/DUO_V6_CEB
if [ "$DEPLOYMENT_GROUP_NAME" == "staging" ]; then
    cp scripts/staging_agent.config /home/DUO_V6_OBSTORE/agent.config
elif ["$DEPLOYMENT_GROUP_NAME" == "production" ]; then
   cp scripts/production_agent.config /home/DUO_V6_OBSTORE/agent.config
fi
