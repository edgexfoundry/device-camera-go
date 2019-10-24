//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

loadGlobalLibrary()

def buildImage_amd64
def buildImage_arm64

def image_amd64
def image_arm64

pipeline {
    agent {
        label 'centos7-docker-4c-2g'
    }

    options {
        timestamps()
    }

    stages {
        stage('LF Prep') {
            steps {
                edgeXSetupEnvironment()
                edgeXSemver 'init'
                script {
                    def semverVersion = edgeXSemver()
                    env.setProperty('VERSION', semverVersion)
                    sh 'echo $VERSION > VERSION'
                    stash name: 'semver', includes: '.semver/**,VERSION', useDefaultExcludes: false
                }
            }
        }

        stage('Multi-Arch Build') {
            // fan out
            parallel {
                stage('Build amd64') {
                    stages {
                        stage('Test') {
                            steps {
                                script {
                                    // prep, ephemeral image is used as a caching layer
                                    buildImage_amd64 = docker.build(
                                        'image-base-build-amd64',
                                        '-f Dockerfile.build --build-arg BASE=nexus3.edgexfoundry.org:10003/edgex-golang-base:1.12.6-alpine .'
                                    )

                                    // test codecov
                                    buildImage_amd64.inside('-u 0:0') {
                                        sh 'make test'

                                        // do not run on sandbox
                                        if(env.SILO && env.SILO != 'sandbox') {
                                            edgeXCodecov('device-camera-go-codecov-token')
                                        }
                                    }

                                    // fix permissions from the -u 0:0 above
                                    sh 'sudo chown -R jenkins:jenkins $WORKSPACE/*'
                                }
                            }
                        }
                        stage('Docker Build') {
                            steps {
                                unstash 'semver'

                                sh 'echo Currently Building version: `cat ./VERSION`'

                                script {
                                    // This is the main docker image that will be pushed
                                    // BASE image = image from above
                                    image_amd64 = docker.build(
                                        'docker-device-camera-go',
                                        "--build-arg BASE=image-base-build-amd64 --label 'git_sha=${env.GIT_COMMIT}' --label version=${env.VERSION} ."
                                    )
                                }
                            }
                        }
                        stage('Docker Push') {
                            when { expression { edgex.isReleaseStream() } }

                            steps {
                                script {
                                    edgeXDockerLogin(settingsFile: env.MVN_SETTINGS)

                                    docker.withRegistry("https://${env.DOCKER_REGISTRY}:10004") {
                                        image_amd64.push(env.VERSION)
                                        image_amd64.push('latest')
                                        image_amd64.push("${env.SEMVER_BRANCH}")
                                        image_amd64.push("${env.GIT_COMMIT}-${env.VERSION}")
                                    }
                                }
                            }
                        }
                    }
                }
                stage('Build arm64') {
                    agent {
                        label 'ubuntu18.04-docker-arm64-4c-2g'
                    }
                    stages {
                        stage('Test') {
                            steps {
                                script {
                                    // prep, ephemeral image is used as a caching layer
                                    buildImage_arm64 = docker.build(
                                        'image-base-build-arm64',
                                        '-f Dockerfile.build --build-arg BASE=nexus3.edgexfoundry.org:10003/edgex-golang-base:1.12.6-alpine-arm64 .'
                                    )

                                    // test codecov
                                    buildImage_arm64.inside('-u 0:0') {
                                        sh 'make test'

                                        // do not run on sandbox
                                        if(env.SILO && env.SILO != 'sandbox') {
                                            edgeXCodecov('device-camera-go-codecov-token')
                                        }
                                    }

                                    // fix permissions from the -u 0:0 above
                                    sh 'sudo chown -R jenkins:jenkins $WORKSPACE/*'
                                }
                            }
                        }
                        stage('Docker Build') {
                            steps {
                                unstash 'semver'

                                sh 'echo Currently Building version: `cat ./VERSION`'

                                script {
                                    // This is the main docker image that will be pushed
                                    // BASE image = image from above
                                    image_arm64 = docker.build(
                                        'docker-device-camera-go-arm64',
                                        "--build-arg BASE=image-base-build-arm64 --label 'git_sha=${env.GIT_COMMIT}' --label version=${env.VERSION} ."
                                    )
                                }
                            }
                        }
                        stage('Docker Push') {
                            when { expression { edgex.isReleaseStream() } }

                            steps {
                                script {
                                    edgeXDockerLogin(settingsFile: env.MVN_SETTINGS)

                                    docker.withRegistry("https://${env.DOCKER_REGISTRY}:10004") {
                                        image_arm64.push(env.VERSION)
                                        image_arm64.push('latest')
                                        image_arm64.push("${env.SEMVER_BRANCH}")
                                        image_arm64.push("${env.GIT_COMMIT}-${env.VERSION}")
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }

        stage('Snyk Scan') {
            when { expression { edgex.isReleaseStream() } }
            steps {
                edgeXSnyk()
            }
        }

        stage('Clair Scan') {
            when { expression { edgex.isReleaseStream() } }
            steps {
                edgeXClair("${env.DOCKER_REGISTRY}:10004/docker-device-camera-go:${env.GIT_COMMIT}-${env.VERSION}")
                edgeXClair("${env.DOCKER_REGISTRY}:10004/docker-device-camera-go-arm64:${env.GIT_COMMIT}-${env.VERSION}")
            }
        }

        stage('SemVer Tag') {
            when { expression { edgex.isReleaseStream() } }
            steps {
                unstash 'semver'
                sh 'echo v${VERSION}'
                edgeXSemver('tag')
                edgeXInfraLFToolsSign(command: 'git-tag', version: 'v${VERSION}')
            }
        }

        stage('Semver Bump Pre-Release Version') {
            when { expression { edgex.isReleaseStream() } }
            steps {
                edgeXSemver('bump pre')
                edgeXSemver('push')
            }
        }
    }

    post {
        failure {
            script {
                currentBuild.result = "FAILED"
            }
        }
        always {
            edgeXInfraPublish()
        }
    }
}

def loadGlobalLibrary(branch = '*/master') {
    library(identifier: 'edgex-global-pipelines@master', 
        retriever: legacySCM([
            $class: 'GitSCM',
            userRemoteConfigs: [[url: 'https://github.com/edgexfoundry/edgex-global-pipelines.git']],
            branches: [[name: branch]],
            doGenerateSubmoduleConfigurations: false,
            extensions: [[
                $class: 'SubmoduleOption',
                recursiveSubmodules: true,
            ]]]
        )
    ) _
}
