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
        timeout(time: 20, unit: 'MINUTES')
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
                sh 'go version'
                sh 'go env'
            }
        }

        stage('Download Dependencies') {
            steps {
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

        stage('Run Tests & Coverage') {
            steps {
                echo 'Running tests with coverage...'
                sh '''
                    mkdir -p coverage
                    go test -v ./... -coverprofile=coverage/coverage.out -covermode=atomic || true
                    go tool cover -html=coverage/coverage.out -o coverage/coverage.html || true
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
                echo 'Running gosec security scanner (errors ignored)...'
                sh '''
                    if ! command -v gosec &> /dev/null; then
                        go install github.com/securego/gosec/v2/cmd/gosec@latest
                    fi
                    gosec ./... -fmt=json -out=gosec-report.json || echo "Gosec failed, but continuing"
                '''
            }
        }

        stage('Generate JUnit Test Report') {
            steps {
                echo 'Generating JUnit XML test report...'
                sh '''
                    if ! command -v go-junit-report &> /dev/null; then
                        go install github.com/jstemmer/go-junit-report@latest
                    fi
                    go test -v ./... 2>&1 | go-junit-report > report.xml || echo "<testsuites></testsuites>" > report.xml
                '''
            }
        }
    }

    post {
        always {
            echo 'Publishing reports and cleaning workspace...'

            junit allowEmptyResults: true, testResults: 'report.xml'

            publishHTML([
                allowMissing: true,
                alwaysLinkToLastBuild: true,
                keepAll: true,
                reportDir: 'coverage',
                reportFiles: 'coverage.html',
                reportName: 'Go Coverage Report'
            ])

            archiveArtifacts artifacts: 'coverage/coverage.out,coverage/coverage.html,report.xml,golangci-lint-report.xml,gosec-report.json', fingerprint: true

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
