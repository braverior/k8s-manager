#!/bin/sh

source ./version.sh


PNAME="k8s-manager-web"

if [ x"$1" == "xtest" ]; then
    PNAME="k8s-manager-web-sandbox"
    echo "build test images ${VERSION}"
fi

API_MASTER_SERVER="k8s-manager-api-master.domob-inc.com"
# 如果有Dockerfile模板，则需要以下代码
sed 's/__VERSION__/'${VERSION}'/' Dockerfile-tpl  > Dockerfile-tmp
sed -i  's/127.0.0.1:30001/'${API_MASTER_SERVER}'/'  vite.config.ts


# 替换旧版dockerfile
mv Dockerfile-tmp Dockerfile

rm -f Dockerfile-tmp
# 开始编译Docker镜像，如果有一些初始化操作亦可以放到这里


################################################################
#  以下代码一般无需更改，如果想配置推送机房，则更新REGIONS即可 #
################################################################

#推送的地域，分别为华北、华东、华南
REGIONS="beijingbd"


echo "docker build..."
docker build -t $PNAME .

if [ $? != 0 ];then
    echo "build ${PNAME} failed."
    exit -1
fi

docker tag $PNAME:latest $PNAME:${VERSION}


# 清理本地的镜像，避免无用镜像过多
function clean() {
	DOCKER_REPO_TMP=$1
	if [ x$1 == x"" ]; then
        echo "not found docker repo argument."
        return -1
    fi
	docker rmi ${DOCKER_REPO_TMP}/${PNAME}:latest
	docker rmi ${DOCKER_REPO_TMP}/${PNAME}:${VERSION}
}

# 推送前打tag
function tag() {
	DOCKER_REPO_TMP=$1
	if [ x$1 == x"" ]; then
		echo "not found docker repo argument."
            return -1
        fi
	echo "docker tag ${DOCKER_REPO_TMP} latest..."
	docker tag ${PNAME}:latest ${DOCKER_REPO_TMP}/${PNAME}:latest
	echo "docker tag ${DOCKER_REPO_TMP} ${VERSION}..."
	docker tag ${PNAME}:latest ${DOCKER_REPO_TMP}/${PNAME}:${VERSION}
}
# 推送镜像，有三次重试
function push() {
	DOCKER_REPO_TMP=$1
	if [ x$1 == x"" ]; then
		echo "not found docker repo argument."
                return -1
        fi

	TAG_TMP=$2
	if [ x$2 == x"" ]; then
		echo "not found docker tag argument."
                return -1
        fi

	echo "starting to push ${PNAME}:${TAG_TMP} to ${DOCKER_REPO_TMP}/${PNAME}:${TAG_TMP}..."
	succ=0
	for i in `seq 1 3`; do
		docker push ${DOCKER_REPO_TMP}/${PNAME}:${TAG_TMP}
		if [ $? != 0 ];then
			echo "[$i times] docker push failed..."
			echo "trying to push ${PNAME}:${TAG_TMP} to ${DOCKER_REPO_TMP}/${PNAME}:${TAG_TMP}..."
		else
			echo "[$i times] docker push is successful..."
			succ=1
			break
		fi
	done

	if [ $succ == 0 ]; then
		echo "failed to push docker images ${DOCKER_REPO_TMP}/${PNAME}:${TAG_TMP} at last..., exit now."
		exit -1
	fi
}



for REGION in $REGIONS;do
  if [ $REGION == "shanghaitx" ]; then
    DOCKER_REPO="domob-hub-hd.tencentcloudcr.com/domob-hub"
  elif [ $REGION == "oversea" ]; then
    DOCKER_REPO="oversea-hub.domob-inc.com/domob-hub"
  elif [ $REGION == "oversea-sg" ]; then
    DOCKER_REPO="domob-hub-sg.tencentcloudcr.com/domob-hub"
  elif [ $REGION == "oversea-va" ]; then
    DOCKER_REPO="domob-hub-va.tencentcloudcr.com/domob-hub"
  elif [ $REGION == "oversea-fra" ]; then
    DOCKER_REPO="domob-hub-fra.tencentcloudcr.com/domob-hub"
  elif [ $REGION == "beijingbd" ]; then
    DOCKER_REPO="ccr-2107uxdo-pub.cnc.bj.baidubce.com/domob-hub"
  else
    DOCKER_REPO="registry.cn-$REGION.aliyuncs.com/domob-hub"
  fi
	echo "runing as region in $REGION.."
	tag $DOCKER_REPO
	if [ $? != 0 ];then
		echo "docker tag failed.."
		exit -1
	fi
	push $DOCKER_REPO latest
	push $DOCKER_REPO ${VERSION}
	clean $DOCKER_REPO
done


echo "cleaning none tag images"

#docker images|grep none|awk '{print $3}'|xargs docker rmi >/dev/null 2>&1

docker images |grep $PNAME

echo "build complete"
