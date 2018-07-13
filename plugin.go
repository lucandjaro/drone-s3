package main

import (
	"mime"
	"os"
	"path/filepath"
	"strings"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mattn/go-zglob"
)

// Plugin defines the S3 plugin parameters.
type Plugin struct {
	Endpoint string
	Key      string
	Secret   string
	Bucket   string
	// if not "", enable server-side encryption
	// valid values are:
	//     AES256
	//     aws:kms
	Encryption string

	// us-east-1
	// us-west-1
	// us-west-2
	// eu-west-1
	// ap-southeast-1
	// ap-southeast-2
	// ap-northeast-1
	// sa-east-1
	Region string

	// Indicates the files ACL, which should be one
	// of the following:
	//     private
	//     public-read
	//     public-read-write
	//     authenticated-read
	//     bucket-owner-read
	//     bucket-owner-full-control
	Access string

	// Copies the files from the specified directory.
	// Regexp matching will apply to match multiple
	// files
	//
	// Examples:
	//    /path/to/file
	//    /path/to/*.txt
	//    /path/to/*/*.txt
	//    /path/to/**
	Source string
	Target string

	// Strip the prefix from the target path
	StripPrefix string

	YamlVerified bool

	// Exclude files matching this pattern.
	Exclude []string

	// Use path style instead of domain style.
	//
	// Should be true for minio and false for AWS.
	PathStyle bool
	// Dry run without uploading/
	DryRun bool
	CreateBucketIfNecessary bool

	AppendBranchtoBucket bool

	s3PrefixStripBranch string

	CommitBranch string

	S3Hosting bool
	IndexDocument string
	ErrorDocument string
}

// Exec runs the plugin
func (p *Plugin) Exec() error {
	if (p.AppendBranchtoBucket == true){
		toAppend := []string{"", ""}
		toAppend[0] = p.Bucket
		if (len(p.s3PrefixStripBranch) > 0){
			toAppend[1] = strings.ToLower(strings.TrimPrefix(p.CommitBranch, p.s3PrefixStripBranch))
		} else{
			toAppend[1] = strings.ToLower(p.CommitBranch)
		}

		p.Bucket = strings.Join(toAppend, "-")
	}

	// Replace _ by - because unsupported by aws
	p.Bucket = strings.Replace(p.Bucket, "_", "-", -1)
	
	// normalize the target URL
	if strings.HasPrefix(p.Target, "/") {
		p.Target = p.Target[1:]
	}
	
	// create the client
	conf := &aws.Config{
		Region:           aws.String(p.Region),
		Endpoint:         &p.Endpoint,
		DisableSSL:       aws.Bool(strings.HasPrefix(p.Endpoint, "http://")),
		S3ForcePathStyle: aws.Bool(p.PathStyle),
	}
	
	//Allowing to use the instance role or provide a key and secret
	if p.Key != "" && p.Secret != "" {
		conf.Credentials = credentials.NewStaticCredentials(p.Key, p.Secret, "")
	}else if p.YamlVerified != true {
		return errors.New("Security issue: When using instance role you must have the yaml verified")
	}
	client := s3.New(session.New(), conf)
	
	// Add Creation of Bucket if needed
	if (p.CreateBucketIfNecessary == true){
		
		input := &s3.ListBucketsInput{
		}
		result, err := client.ListBuckets(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					log.WithFields(log.Fields{
						"err": aerr.Error(),
					}).Error("ListBucket")
					return err
					
					//fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				//fmt.Println(err.Error())
				log.WithFields(log.Fields{
					"err": aerr.Error(),
				}).Error("ListBucket")
				return err
			}
		}

		// log.WithFields(log.Fields{
		// 	"listBucket": result,
		// }).Info("ListBuckets")
		
		var isBucketExisting bool = false
		
		for _, bucket := range result.Buckets {
			// log.WithFields(log.Fields{
			// 	"BucketName": p.Bucket,
			// 	"BucketNamefromAWS": aws.StringValue(bucket.Name),
			// }).Info("AWS Check\n")
			if ((aws.StringValue(bucket.Name) == p.Bucket) == true){
				isBucketExisting = true
			}
		}
		
		log.WithFields(log.Fields{
			"isBucketExisting": isBucketExisting,
		}).Info("isBucketExisting")
		
		if (isBucketExisting == false){
			_, err = client.CreateBucket(&s3.CreateBucketInput{
				Bucket: aws.String(p.Bucket),
			})
			if err != nil {
				log.WithFields(log.Fields{
					"BucketName": p.Bucket,
					"Err": err,
				}).Error("Unable to create bucket")
				return err
			}
			
			// Wait until bucket is created before finishing
			log.WithFields(log.Fields{
				"BucketName": p.Bucket,
			}).Info("Waiting for bucket to be created...\n")
			
			err = client.WaitUntilBucketExists(&s3.HeadBucketInput{
				Bucket: aws.String(p.Bucket),
			})
			if err != nil {
				log.WithFields(log.Fields{
					"BucketName": p.Bucket,
					"Err": err,
				}).Error("Error occurred while waiting for bucket to be created")
				return err
			}
			
			log.WithFields(log.Fields{
				"BucketName": p.Bucket,
			}).Info("Bucket created\n")
			
		}
	}


	if (p.S3Hosting == true){
		// S3 Config of Bucket
		result, err := client.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(p.Bucket),
		})
		if err != nil {
			// Check for the NoSuchWebsiteConfiguration error code telling us
			// that the bucket does not have a website configured.
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchWebsiteConfiguration" {
				log.WithFields(log.Fields{
					"BucketName": p.Bucket,
					"Err": err,
				}).Error("Bucket does not have website configuration")
			}
		}

		log.WithFields(log.Fields{
			"BucketConfig": result,
		}).Info("Bucket Website Configuration")
		
		params := s3.PutBucketWebsiteInput{
			Bucket: aws.String(p.Bucket),
			WebsiteConfiguration: &s3.WebsiteConfiguration{},
		}
		
		if (len(p.IndexDocument) > 0 ){
			// Add the index page if set on CLI
			params.WebsiteConfiguration.IndexDocument = &s3.IndexDocument{
				Suffix: aws.String(p.IndexDocument),
			}
			
		// Add the error page if set on CLI
			if (len(p.ErrorDocument) > 0){
				params.WebsiteConfiguration.ErrorDocument = &s3.ErrorDocument{
					Key: aws.String(p.ErrorDocument),
				}  
			}
		}

		_, err = client.PutBucketWebsite(&params)
		if err != nil {
			
			log.WithFields(log.Fields{
				"BucketName": p.Bucket,
				"Err": err,
			}).Error("Unable to set bucket website configuration")
		}else{
		
			log.WithFields(log.Fields{
				"BucketName": p.Bucket,
			}).Info("Successfully set bucket website configuration")
			
			s := []string{}
			s = append(s, p.Bucket)
		s = append(s, ".s3-website.eu-west-3.amazonaws.com")
			log.WithFields(log.Fields{
				"S3 URL": strings.Join(s, ""),
			}).Info("S3 URL is available")
		}
	}
		

	// End of Creation of Bucket	
	matches, err := matches(p.Source, p.Exclude)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Could not match files")
		return err
	}

	for _, match := range matches {

		stat, err := os.Stat(match)
		if err != nil {
			continue // should never happen
		}

		// skip directories
		if stat.IsDir() {
			continue
		}

		target := filepath.Join(p.Target, strings.TrimPrefix(match, p.StripPrefix))
		if !strings.HasPrefix(target, "/") {
			target = "/" + target
		}

		// amazon S3 has pretty crappy default content-type headers so this pluign
		// attempts to provide a proper content-type.
		content := contentType(match)

		// log file for debug purposes.
		log.WithFields(log.Fields{
			"name":         match,
			"bucket":       p.Bucket,
			"target":       target,
			"content-type": content,
		}).Info("Uploading file")

		// when executing a dry-run we exit because we don't actually want to
		// upload the file to S3.
		//if p.DryRun {
		//	continue
		//}

		f, err := os.Open(match)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"file":  match,
			}).Error("Problem opening file")
			return err
		}
		defer f.Close()

		putObjectInput := &s3.PutObjectInput{
			Body:        f,
			Bucket:      &(p.Bucket),
			Key:         &target,
			ACL:         &(p.Access),
			ContentType: &content,
		}

		if p.Encryption != "" {
			putObjectInput.ServerSideEncryption = &(p.Encryption)
		}

		_, err = client.PutObject(putObjectInput)

		// log file for debug purposes.
		log.WithFields(log.Fields{
			"name":         match,
			"bucket":       p.Bucket,
			"target":       target,
			"content-type": content,
		}).Info("Uploaded file")

		
		
		if err != nil {
			log.WithFields(log.Fields{
				"name":   match,
				"bucket": p.Bucket,
				"target": target,
				"error":  err,
			}).Error("Could not upload file")
			
			return err
		}
		f.Close()
	}
	
	return nil
}

// matches is a helper function that returns a list of all files matching the
// included Glob pattern, while excluding all files that matche the exclusion
// Glob pattners.
func matches(include string, exclude []string) ([]string, error) {
	matches, err := zglob.Glob(include)
	if err != nil {
		return nil, err
	}
	if len(exclude) == 0 {
		return matches, nil
	}

	// find all files that are excluded and load into a map. we can verify
	// each file in the list is not a member of the exclusion list.
	excludem := map[string]bool{}
	for _, pattern := range exclude {
		excludes, err := zglob.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range excludes {
			excludem[match] = true
		}
	}

	var included []string
	for _, include := range matches {
		_, ok := excludem[include]
		if ok {
			continue
		}
		included = append(included, include)
	}
	return included, nil
}

// contentType is a helper function that returns the content type for the file
// based on extension. If the file extension is unknown application/octet-stream
// is returned.
func contentType(path string) string {
	ext := filepath.Ext(path)
	typ := mime.TypeByExtension(ext)
	if typ == "" {
		typ = "application/octet-stream"
	}
	return typ
}
