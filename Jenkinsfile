pipeline {
    agent any

    environment {
        GO111MODULE = 'on'
        CGO_ENABLED = '0'
        GOPATH = "${WORKSPACE}/go"
        GOCACHE = "${GOPATH}/pkg/mod/cache"
        PATH = "/opt/homebrew/bin:${GOPATH}/bin:${env.PATH}"
    }

    options {
        // Limitar el tiempo máximo de ejecución del pipeline
        timeout(time: 20, unit: 'MINUTES')
        // Mantener solo 10 builds para no llenar disco
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }

    stages {
        stage('Checkout') {
            steps {
                echo 'Cloning repository...'
                checkout scm
            }
        }

        stage('Setup Go Environment') {
            steps {
                echo 'Go version and environment'
                sh '''
                    go version
                    go env
                '''
            }
        }

        stage('Download Dependencies') {
            steps {
                echo 'Downloading Go modules...'
                sh '''
                    go mod download
                    go mod verify
                '''
            }
        }

        stage('Build') {
            steps {
                echo 'Building application...'
                sh 'go build -v ./...'
            }
        }

        stage('Parallel Analysis') {
            parallel {
                stage('Run Tests') {
                    steps {
                        echo 'Running tests with coverage...'
                        sh '''
                            go test -v ./... -coverprofile=coverage.out -covermode=atomic
                            go tool cover -html=coverage.out -o coverage.html
                        '''
                    }
                }

                stage('Code Quality - Lint') {
                    steps {
                        echo 'Running golangci-lint...'
                        sh '''
                            if ! command -v golangci-lint &> /dev/null; then
                                curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin
                            fi
                            golangci-lint run --out-format checkstyle > golangci-lint-report.xml || true
                        '''
                    }
                }

                stage('Security Scan') {
                    steps {
                        echo 'Running gosec security scanner...'
                        sh '''
                            if ! command -v gosec &> /dev/null; then
                                go install github.com/securego/gosec/v2/cmd/gosec@latest
                            fi
                            gosec -fmt=json -out=gosec-report.json ./... || true
                        '''
                    }
                }
            }
        }

        stage('Generate Test Reports') {
            steps {
                echo 'Converting test results to JUnit format...'
                sh '''
                    if ! command -v go-junit-report &> /dev/null; then
                        go install github.com/jstemmer/go-junit-report@latest
                    fi
                    go test -v ./... 2>&1 | go-junit-report > report.xml
                '''
            }
        }
    }

    post {
        always {
            echo 'Publishing reports and cleaning workspace...'

            // Test reports
            junit 'report.xml'

            // HTML coverage
            publishHTML([
                allowMissing: false,
                alwaysLinkToLastBuild: true,
                keepAll: true,
                reportDir: '.',
                reportFiles: 'coverage.html',
                reportName: 'Go Coverage Report'
            ])

            // Archive all reports
            archiveArtifacts artifacts: 'coverage.out,coverage.html,report.xml,golangci-lint-report.xml,gosec-report.json', fingerprint: true

            // Clean workspace
            cleanWs()
        }

        success {
            echo '✅ Pipeline completed successfully!'
        }

        failure {
            echo '❌ Pipeline failed. Check the logs above.'
        }
    }
}
