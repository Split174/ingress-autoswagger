node('dockerhost') {
    env.DOCKER_IMAGE = 'docker-devops.art.lmru.tech/bricks/ingress-autoswagger'
    env.DOCKER_REGISTRY_CREDS = 'LM-sa-devops'

    timestamps {
        ansiColor('xterm') {
            stage('Checkout') {
                checkout scm
            }

            stage('Build & Push Image') {
                if (env.CHANGE_ID) {
                    lint()
                } else {
                    image_build_and_push(docker_image_name)
                }

            }

            stage('Wipe') {
                cleanWs()
            }
        }
    }
}

def lint() {
    // not needed here
}

def image_build_and_push(docker_image_name) {
    def image = docker.build("${env.DOCKER_IMAGE}:2.0", ".")
    try {
        docker.withRegistry("https://$DOCKER_IMAGE", "$DOCKER_REGISTRY_CREDS") {
            image.push('2.0')
        }
    }
    finally {
        sh "docker rmi $DOCKER_IMAGE:2.0"
    }
}
