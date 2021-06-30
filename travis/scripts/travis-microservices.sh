#!/bin/bash

gererateMiddlewarePermissions()
{
  # # LONG FORM GENERATE: 
  # # install deps and generate the permisssions.go used by micro-service middleware
  # cd auth-service
  # sh -x install_dependencies.sh
  # bash ./deployment/build.sh --env dev --cloud gcloud

  # SHORT FORM GENERATE: 
  cd auth-service
  npm install --prefix deployment
  npm i -g typescript@3.4.5
  tsc deployment/src/permissions-only.ts
  node deployment/src/permissions-only.js
  ls authorization/middleware/permissions
  cd ..
}

beforeInstall()
{
  gererateMiddlewarePermissions
  chmod +x checkmarx/pullgit.sh
 # export GOBIN=${GOPATH}/bin
 # export GOPATH=${TRAVIS_BUILD_DIR}
  echo ${GOPATH}
  echo ${TRAVIS_BUILD_DIR}
 # export GOBIN="$EFFECTIVE_GOPATH/bin"
 # echo ${GOBIN}
 # export PATH="${PATH}:${GOBIN}"
  sudo apt-get install -y libxml2
  sudo apt-get install -y libxml2-dev
  sudo apt-get install -y pkg-config
  apt install librdkafka-dev
  # export PKG_CONFIG_PATH=${PKG_CONFIG_PATH}:/usr/lib/pkgconfig
  # git clone https://github.com/edenhill/librdkafka.git
  # cd librdkafka
  # ./configure --prefix /usr
  # make
  # sudo make install
  # sudo apt-get update && sudo apt-get install -y software-properties-common
  # sudo add-apt-repository ppa:0k53d-karl-f830m/openssl -y
  # sudo apt-get update
  sudo apt-get install openssl
  wget https://github.com/neo4j-drivers/seabolt/releases/download/v1.7.3/seabolt-1.7.3-Linux-ubuntu-16.04.deb
  sudo dpkg -i seabolt-1.7.3-Linux-ubuntu-16.04.deb
  sudo rm seabolt-1.7.3-Linux-ubuntu-16.04.deb
  # cd ${TRAVIS_BUILD_DIR}
  # Setup dependency management tool
  curl -L -s https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64 -o $GOPATH/bin/dep
  chmod +x $GOPATH/bin/dep
  dep ensure -vendor-only

  download_url=$(curl -s https://api.github.com/repos/go-swagger/go-swagger/releases/tags/v0.23.0 | \
	jq -r '.assets[] | select(.name | contains("'"$(uname | tr '[:upper:]' '[:lower:]')"'_amd64")) | .browser_download_url')
  sudo curl -o /usr/local/bin/swagger -L "${download_url}"
  sudo chmod +x /usr/local/bin/swagger
}

checkVersion()
{
  if [ -z "$version" ]
  then
    echo "No version found in environment variable, using default: latest"
    version=latest
  else
    echo "version environment variable found, using $version"
  fi
}

run()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
  cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
  make service-docker build=all || exit 1;
  #  kill %1;
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
  checkVersion
  # make push-icr version=$version;
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
    # make push-dockers version=$version;
  #  fi

}

pushAllDockerImages()
{
  cd integration-tests/;
  checkVersion
  make push-to-icr version=$version;
}

scanTwistlock()
{
  CONTAINER=$1
  #sudo curl -k -ssl -u ${TL_USER}:${TL_PASS} --output /tmp/twistcli ${TL_CONSOLE_URL}/api/v1/util/twistcli && sudo chmod a+x /tmp/twistcli && sudo /tmp/twistcli images scan $CONTAINER --details -address ${TL_CONSOLE_URL} -u ${TL_USER} -p ${TL_PASS} || exit 1
}

runAdministrationService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=administration-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/administration-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-administration-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
     checkVersion

  if [[ $1 == "push" ]]
  then
    make push-administration-service-dockers version=$version
  fi
     # make push-administration-service-dockers version=$version;
  #  fi

}

runAuthService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=auth-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/auth-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-auth-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-auth-service-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-auth-service-dockers version=$version
  fi

}

runAnchorService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=anchor-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/anchor-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-anchor-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-anchor-service-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-anchor-service-dockers version=$version
  fi

}

runApiService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=api-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/api-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-api-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-api-service-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-api-service-dockers version=$version
  fi

}

runCallbackService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=callback-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/callback-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-callback-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-callback-service-dockers version=$version;
  #  fi

}

runCryptoService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=crypto-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/crypto-service:latest"
  #  scanTwistlock "gftn/crypto-service-prod:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-crypto-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-crypto-service-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-crypto-service-dockers version=$version
  fi

}

runFeeService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=fee-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/fee-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-fee-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-fee-service-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-fee-service-dockers version=$version
  fi

}

runGasService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=gas-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/gas-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-gas-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-gas-service-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-gas-service-dockers version=$version
  fi

}

runGlobalWhitelistService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=global-whitelist-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/global-whitelist-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-global-whitelist-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-global-whitelist-service-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-global-whitelist-service-dockers version=$version
  fi

}

runParticipantRegistry()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=participant-registry || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/participant-registry:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-participant-registry-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-participant-registry-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-participant-registry-dockers version=$version
  fi

}

runPaymentListener()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=payment-listener || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/payment-listener:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-payment-listener-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-payment-listener-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-payment-listener-dockers version=$version
  fi
}

runPayoutService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=payout-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/payout-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-payout-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
     # make push-payout-service-dockers version=$version;
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-payout-service-dockers version=$version
  fi

}

runQuotesService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=quotes-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/quotes-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-quotes-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-quotes-service-dockers version=$version
  fi
}

runSendService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=send-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/send-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-send-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-send-service-dockers version=$version
  fi
}


runWWGateway()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=ww-gateway || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/ww-gateway:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-ww-gateway-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
  #  fi
  if [[ $1 == "push" ]]
  then
     make push-ww-gateway-dockers version=$version
  fi

}

runAutomationService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=automation-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/automation-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-automation-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-automation-service-dockers version=$version
  fi

}

runPortalAPIService()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=portal-api-service || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/auth-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-auth-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-portal-api-service-dockers version=$version
  fi

}

runWWPortal()
{

  #  gitBranch=`git branch | grep \* | cut -d " " -f2`
   # echo ${gitBranch}
  #  git status
  #  make swaggergen
  #  version=`cat VERSION`
   cd integration-tests/;
  #  while sleep 5m; do echo "=====[ $SECONDS seconds, docker images are still building... ]====="; done &
   make service-docker build=world-wire-web || exit 1;
  #  kill %1;
  #  scanTwistlock "gftn/auth-service:latest"
  #  if [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "development" ]; then
    #  make push-auth-service-dockers version="latest";
  #  elif [ ${TRAVIS_EVENT_TYPE} == 'cron' ] && [ ${TRAVIS_BRANCH} == "publish-docker-images" ]; then
  checkVersion
  #  fi
  if [[ $1 == "push" ]]
  then
    make push-ww-portal-dockers version=$version
  fi

}

afterFailure()
{

  # dump the last 2000 lines of our build to show error
  tail -n 2000 build.log

}

afterSuccess()
{

  # Log that the build was a success
  echo "build succeeded.."

}

CMD=$1
OPT=$2

if [ $CMD = "beforeInstall" ]; then
  beforeInstall
elif [ $CMD = "run" ]; then
  run
elif [ $CMD = "runAdministrationService" ]; then
  runAdministrationService $OPT
elif [ $CMD = "runAnchorService" ]; then
  runAnchorService $OPT
elif [ $CMD = "runApiService" ]; then
  runApiService $OPT
elif [ $CMD = "runAuthService" ]; then
  runAuthService $OPT
elif [ $CMD = "runAutomationService" ]; then
  runAutomationService $OPT
elif [ $CMD = "runCallbackService" ]; then
  runCallbackService $OPT
elif [ $CMD = "runCryptoService" ]; then
  runCryptoService $OPT
elif [ $CMD = "runFeeService" ]; then
  runFeeService $OPT
elif [ $CMD = "runGasService" ]; then
  runGasService $OPT
elif [ $CMD = "runGlobalWhitelistService" ]; then
  runGlobalWhitelistService $OPT
elif [ $CMD = "runParticipantRegistry" ]; then
  runParticipantRegistry $OPT
elif [ $CMD = "runPaymentListener" ]; then
  runPaymentListener $OPT
elif [ $CMD = "runPayoutService" ]; then
  runPayoutService $OPT
elif [ $CMD = "runQuotesService" ]; then
  runQuotesService $OPT
elif [ $CMD = "runSendService" ]; then
  runSendService $OPT
elif [ $CMD = "runWWGateway" ]; then
  runWWGateway $OPT
elif [ $CMD = "runPortalAPIService" ]; then
  runPortalAPIService $OPT
elif [ $CMD = "runWWPortal" ]; then
  runWWPortal $OPT
elif [ $CMD = "pushAllDockerImages" ]; then
 pushAllDockerImages
elif [ $CMD = "afterFailure" ]; then
 afterFailure
elif [ $CMD = "afterSuccess" ]; then
 afterSuccess
fi