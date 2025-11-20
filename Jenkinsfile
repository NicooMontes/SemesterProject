pipeline {
    agent any

    environment {
        GO111MODULE = 'on'
        CGO_ENABLED = '0'
        GOPATH = "${WORKSPACE}/go"
        PATH = "/opt/homebrew/bin:${GOPATH}/bin:${env.PATH}"
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
                echo 'Setting up Go environment...'
                sh '''
                    go version
                    go env
                '''
            }
        }

        stage('Download Dependencies') {
            steps {
                echo 'Downloading Go modules...'
                sh 'go mod download'
                sh 'go mod verify'
            }
        }

        stage('Build') {
            steps {
                echo 'Building application...'
                sh 'go build -v ./...'
            }
        }

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
                    # Install golangci-lint if not present
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
                    # Install gosec if not present
                    if ! command -v gosec &> /dev/null; then
                        go install github.com/securego/gosec/v2/cmd/gosec@latest
                    fi
                    gosec -fmt=json -out=gosec-report.json ./... || true
                '''
            }
        }

        stage('Generate Test Reports') {
            steps {
                echo 'Converting test results to JUnit format...'
                sh '''
                    # Install go-junit-report if not present
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
            // Publish test results
            junit 'report.xml'

            // Publish coverage report
            publishHTML([
                allowMissing: false,
                alwaysLinkToLastBuild: true,
                keepAll: true,
                reportDir: '.',
                reportFiles: 'coverage.html',
                reportName: 'Go Coverage Report'
            ])

            // Archive artifacts
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
