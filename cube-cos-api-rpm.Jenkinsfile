pipeline {
    agent {
        docker {
            image 'localhost:5000/cube-cos-api-rpm'
            args '-u 1000:1000'
            label 'bldsrv_prod'
        }
    }

    environment {
        OWNER          = 'bigstack-oss'
        PROJ_NAME      = 'cube-cos-api'
        REPO_NAME      = "${OWNER}/${PROJ_NAME}"

        BLDSRV         = 'bldsrv_prod'

        GITHUB_PAT     = 'github-PAT'
        SLACK_CHANNEL  = "#${PROJ_NAME}-ci"
    }

    options {
        timeout(time: 5, unit: 'MINUTES')
        ansiColor('xterm')
    }

    stages {
        stage('init') {
            steps {
                echo 'Initializing the pipeline...'

                script {
                    env.getEnvironment().each { name, value ->
                        println "Name: $name -> Value $value"
                    }
                }

                echo 'Setting up the git environment...'
                sh """
                git config --global --add safe.directory \$(pwd)
                git remote set-url origin git@github.com:${REPO_NAME}.git

                mkdir -p ~/.ssh
                ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
                chmod 600 ~/.ssh/known_hosts
                """


                script {
                    env.BLDPTH = sh(script: 'pwd', returnStdout: true).trim()
                }
            }
        }

        stage('release') {
            steps {
                dir("${env.BLDPTH}") {
                    echo 'Adjusting the spec file due to container limitations...'
                    sh 'sed -i "s;%{_unitdir};/usr/lib/systemd/system;g" ./init/cube-cos-api.spec'
                }

                dir("${env.BLDPTH}") {
                    echo 'Creating the rpm package...'
                    sh 'go-task rpm:build'
                }

                dir("${env.BLDPTH}") {
                    echo 'Uploading the rpm to the GitHub release...'

                    script {
                        def rpms = sh(
                            script: 'ls ~/rpmbuild/RPMS/x86_64/',
                            returnStdout: true
                        ).trim().split(" ")
                        def rpm = ""
                        if (rpms.size() > 0) {
                            rpm = rpms[0]
                        }

                        if (rpm != "") {
                            withCredentials([string(credentialsId: env.GITHUB_PAT, variable: 'PAT')]) {
                                sh 'echo ' + PAT + ' | gh auth login --with-token && ' +
                                    "gh release upload ${env.TAG_NAME} ~/rpmbuild/RPMS/x86_64/${rpm} --repo ${env.REPO_NAME}"
                            }
                        }
                    }
                }

                dir("${env.BLDPTH}") {
                    echo 'Cleaning up the rpm build directory...'
                    sh 'go-task rpm:cleanRpmBuild'
                }
            }
        }
    }

    post {
        always {
            echo 'Cleaning up...'
            cleanWs()

            script {
                echo 'Sending Slack notification...'
                sendSlackNotification(currentBuild.result ?: 'SUCCESS', env.TAG_NAME)
            }
        }
    }
}

def sendSlackNotification(String buildStatus, String version) {
    def color = buildStatus == 'SUCCESS' ? 'good' : 'danger'
    def message = buildStatus == 'SUCCESS' ?
        "Pipeline *${env.JOB_NAME}* completed successfully." :
        "Pipeline *${env.JOB_NAME}* failed. Check logs: ${env.BUILD_URL}"

    if (buildStatus == 'SUCCESS') {
        message += "\n\n<https://github.com/${env.REPO_NAME}/releases/tag/${version} | Release ${version}>"
    }

    slackSend(channel: env.SLACK_CHANNEL, color: color, message: message)
}
