# Environment variable
ARCH="$(uname -m)"
OS_NAME="$(uname -s)"
FABLET_PKG="github.com/IBM/fablet/main"
RELEASE_TARGET="./release/${OS_NAME}_${ARCH}"
FABLET_BIN="fablet"
GO_BUILD_CMD="go build"
WEB_BUILD_CMD="yarn build"
WEB_INSTALL_PKG_CMD="yarn install"
WEB_SRC="./web"
WEB_SRC_BUILD="${WEB_SRC}/build"
WEB_TARGET="${RELEASE_TARGET}/web"
SERVICE="service"
WEB="web"

mkReleaseFolder() {
    mkdir -p "${RELEASE_TARGET}"
}

buildService() {
    echo "Compile Fablet binary files."
    mkReleaseFolder
    rm -f "${RELEASE_TARGET}/${FABLET_BIN}"
    ${GO_BUILD_CMD} -o "${RELEASE_TARGET}/${FABLET_BIN}" ${FABLET_PKG}
}

buildWeb() {
    echo "Compile Fablet web files."
    mkReleaseFolder
    rm -rf "${WEB_TARGET}"
    cd "${WEB_SRC}" && ${WEB_INSTALL_PKG_CMD} && ${WEB_BUILD_CMD} && cd .. && mv "${WEB_SRC_BUILD}" "${WEB_TARGET}"
}


build1=$1
build2=$2
if [ "${build1}" == "${SERVICE}" ] || [ "${build2}" == "${SERVICE}" ]; then
    buildService
    if [ $? -ne 0 ]; then
      echo "Error!"
      exit 1
    fi
fi
if [ "${build1}" == "${WEB}" ] || [ "${build2}" == "${WEB}" ]; then
    buildWeb
    if [ $? -ne 0 ]; then
      echo "Error!"
      exit 1
    fi
fi
if [ -z "${build1}" ] && [ -z "${build2}" ]; then
    buildService
    buildWeb
fi

echo "Please run ${RELEASE_TARGET}/${FABLET_BIN}" to start Fablet.