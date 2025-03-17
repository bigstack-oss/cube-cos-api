pipeline {
    agent {
        docker {
            image 'localhost:5000/cube-cos-api-jail'
            args '-u 1000:1000'
            label 'bldsrv_prod'
        }
    }

    environment {
        OWNER          = 'bigstack-oss'
        PROJ_NAME      = 'cube-cos-api'
        REPO_NAME      = "${OWNER}/${PROJ_NAME}"
        GIT_BRANCH     = "${env.BRANCH_NAME}"

        BLDSRV         = 'bldsrv_prod'

        GITHUB_PAT     = 'github-PAT'
        GITHUB_SSH_KEY = 'github-SSH-KEY'
        SLACK_CHANNEL  = "#${PROJ_NAME}-ci"
    }

    options {
        timeout(time: 5, unit: 'MINUTES')
        ansiColor('xterm')
        lock("${env.JOB_NAME}")
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
                    env.VERSION = getVersion()
                    env.PR_OR_COMMIT = sh(
                        script: 'echo "$(git log -1 --pretty=%B) [$(git log -1 --pretty=format:\'%an\')]"',
                        returnStdout: true
                    ).trim()
                }
            }
        }

        stage('check') {
            steps {
                dir("${env.BLDPTH}") {
                    echo 'Running checks...'
                    sh 'go-task check'
                }
            }
        }

        stage('build') {
            steps {
                dir("${env.BLDPTH}") {
                    echo 'Building the project...'
                    sh 'go-task build'
                }
            }
        }

        stage('test') {
            steps {
                dir("${env.BLDPTH}") {
                    echo 'Running tests...'
                    sh 'go-task test'
                }
            }
        }

        stage('release') {
            when {
                branch 'develop'
            }

            steps {
                dir("${env.BLDPTH}") {
                    echo 'Releasing the binary to GitHub...'

                    script {
                        def version = env.VERSION
                        echo "Generated version: ${version}"

                        sshagent([GITHUB_SSH_KEY]) {
                            script {
                                sh "git tag ${version}"
                                sh "git push origin ${version}"
                            }
                        }

                        def commitish = sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
                        createGitHubRelease(
                            credentialId: GITHUB_PAT,
                            repository: REPO_NAME,
                            tag: version,
                            commitish: commitish,
                            bodyText: "Release ${version}",
                            name: version
                        )
                    }
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
                sendSlackNotification(env.GIT_BRANCH, currentBuild.result ?: 'SUCCESS', env.VERSION, env.PR_OR_COMMIT)
            }
        }
    }
}

def getVersion() {
    def baseVersion = sh(script: 'cat VERSION', returnStdout: true).trim()
    def commitish = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
    
    // hardcode the dev version for now, should be replaced with a proper versioning scheme
    // when we have a release.
    return "${baseVersion}-dev-${commitish}"
}

def sendSlackNotification(String branch, String buildStatus, String version, String prOrCommit) {
    def color = buildStatus == 'SUCCESS' ? 'good' : 'danger'
    def message = buildStatus == 'SUCCESS' ?
        "Pipeline *${env.JOB_NAME}* completed successfully." :
        "Pipeline *${env.JOB_NAME}* failed. Check logs: ${env.BUILD_URL}"
    
    if (buildStatus == 'SUCCESS' && branch == 'develop') {
        message += "\n\n<https://github.com/${env.REPO_NAME}/releases/tag/${version} | Release ${version}>"
    }

    message += "\n\n${prOrCommit}"
    slackSend(channel: env.SLACK_CHANNEL, color: color, message: message)
}
