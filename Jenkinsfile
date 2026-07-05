// First real adopter of the homelabpipe shared library in this homelab (every
// other repo still runs its own long inline pipeline). If a stage here
// behaves unexpectedly, ebook-reader's Jenkinsfile has the equivalent
// hand-written pipeline as a fallback reference.
//
// buildWithKaniko clones this repo over a plain, unauthenticated
// `https://github.com/<user>/<repo>.git` URL, so this repository must be
// public on GitHub.
@Library('homelabpipe') _

homelabpipePipeline {
    image {
        tag = "v${env.BUILD_NUMBER}"
    }
}
