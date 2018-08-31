# Maxwell

A go library and toolkit to declaratively create image transformations on jpg images in an AWS S3 bucket.

## Getting Started

Maxwell comes with a command line tool to resize all of the images in a bucket, and an AWS lambda function which you can deploy
on your own account and configure to listen to CreateItem events on the target bucket and prefix of your choice.
### Prerequisites

1. An AWS Account
2. An AWS S3 Bucket.
3. A target prefix directory in your s3 bucket.
4. A place for output to be placed which is OUTSIDE of your target directory.
5. For the Command line tool, you will need to have AWS credentials setup in your environment.

### Installing

Run `go get github.com/SteveCastle/maxwell`
 example

## Running the tests

Run `go test` from inside the project directory.

## Deployment

Add additional notes about how to deploy this on a live system

## Built With

* [Dropwizard](http://www.dropwizard.io/1.0.2/docs/) - The web framework used
* [Maven](https://maven.apache.org/) - Dependency Management
* [ROME](https://rometools.github.io/rome/) - Used to generate RSS Feeds

## Contributing

Please read [CONTRIBUTING.md](https://gist.github.com/PurpleBooth/b24679402957c63ec426) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags). 

## Authors

* **Stephen Castle* - *Initial work* - [SteveCastle](https://github.com/SteveCastle)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

* This project forks the awesome primitive project to provide a way to create awesome blurred SVG transformations.

