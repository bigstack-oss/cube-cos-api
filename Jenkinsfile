parameters([
    string(defaultValue: 'main', name: 'GIT_BRANCH', trim: true)
])
def BLDSRV = "bldsrv_prod"
def PROJ_NAME = "cube-api"
def BLDPTH = "/home/jenkins/workspace/${JOB_NAME}/${PROJ_NAME}"
def GIT_BRANCH_NAME = ""
def SLACK_CHANNEL="#${PROJ_NAME}"
env.getEnvironment().each { name, value -> println "Name: $name -> Value $value" }
lock("${JOB_NAME}-${BLDSRV}") {
    node("${BLDSRV}") {
        ansiColor('xterm') {
            stage('source') {
                if ( GIT_BRANCH.contains('origin/') ) {
                    GIT_BRANCH_NAME = GIT_BRANCH.replace('origin', '*').trim()
                } else {
                    GIT_BRANCH_NAME = '*/' + GIT_BRANCH
                }
                echo "GIT_BRANCH_NAME = ${GIT_BRANCH_NAME}"
                
                try {
                    checkout([$class: 'GitSCM', branches: [[name: "${GIT_BRANCH_NAME}"]], browser: [$class: 'GithubWeb', repoUrl: 'https://github.com/bigstack-oss/cube-cos-api/'], doGenerateSubmoduleConfigurations: false, extensions: [[$class: 'GitLFSPull'], [$class: 'RelativeTargetDirectory', relativeTargetDir: "${PROJ_NAME}"]], userRemoteConfigs: [[url: "git@github.com:bigstack-oss/${PROJ_NAME}.git"]]])
                } catch (e) {
                    echo "Failed to download repo., remove ${PROJ_NAME} source folder and try again!"
                    sh "sudo rm -rf ${PROJ_NAME}"
                    checkout([$class: 'GitSCM', branches: [[name: "${GIT_BRANCH_NAME}"]], browser: [$class: 'GithubWeb', repoUrl: 'https://github.com/bigstack-oss/cube-cos-api/'], doGenerateSubmoduleConfigurations: false, extensions: [[$class: 'GitLFSPull'], [$class: 'RelativeTargetDirectory', relativeTargetDir: "${PROJ_NAME}"]], userRemoteConfigs: [[url: "git@github.com:bigstack-oss/${PROJ_NAME}.git"]]])
                }
            }
        }
    }
}