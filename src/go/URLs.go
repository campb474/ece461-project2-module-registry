package main

// package imports

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"log"
	// "reflect"

	"github.com/estebangarcia21/subprocess"
	// "subprocess"
	// "fmt"
	// "log"
	// "os"
	// "github.com/go-git/go-git/v5"
	// "time"
)

// * START OF REPO STRUCTS * \\

// Struct for a repository
// Includes each metric, and the total score at the end
// this repo struct will be the input to the linked lists,
// where we will pass urls by accessing the repo's url
type repo struct {
	URL                  string
	repoName			 string
	responsiveness       float64
	correctness          float64
	rampUpTime           float64
	busFactor            float64
	licenseCompatibility float64
	netScore             float64
	next                 *repo
}

// this is a function to utilize createing a new repo and initializing each metric within
func newRepo(url string) *repo {


	r := repo{URL: url}
	r.repoName =(cloneRepo(url))
	r.busFactor = -1
	r.correctness = -1
	r.licenseCompatibility = -1
	r.rampUpTime = -1
	r.netScore = -1

	make_shortlog_file("ECE461ProjectCLI")
	r.busFactor = getBusFactor(r.repoName)


	// r.correctness = getCorrectness(r.URL)
	// r.licenseCompatibility = getLicenseCompatibility(r.URL)
	// r.rampUpTime = getRampUpTime(r.URL)
	// r.responsiveness = getResponsiveness(r.URL)
	//r.totalScore = r.busFactor + int(r.correctness*20) + r.licenseCompatibility + r.rampUpTime + r.responsiveness

	// s := subprocess.New("rmdir --ignore-fail-on-non-empty " + r.repoName, subprocess.Shell)
	// s.Exec()
	clearRepoFolder()

	return &r
}

// * END OF REPO STRUCTS * \\

// * START OF RESPONSIVENESS * \\

// Function to get responsiveness metric score
func getResponsiveness(url string) float64 {
	var command string
	command = "python3 src/python/API.py \"" + url + "\" >> src/python/responsiveness/score.txt"
	s := subprocess.New(command, subprocess.Shell)
	s.Exec()
	file, _ := os.Open("src/python/responsiveness/score.txt")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()
		score, err := strconv.ParseFloat(s, 64)
		if err != nil {
			removeScores()
			return 0
		} else {
			removeScores()
			return score
		}
	}
	removeScores()
	return 0
}

func removeScores() {
	var command string
	command = "python3 -c 'import os; os.remove(\"src/python/responsiveness/score.txt\") if os.path.exists(\"src/python/responsiveness/score.txt\") else \"continue\";'"
	r := subprocess.New(command, subprocess.Shell)
	r.Exec()
}

// * END OF RESPONSIVENESS * \\

// * START OF RAMP-UP TIME * \\

// Function to get ramp-up time metric scor
func getRampUpTime(url string) float64 {
	var command string
	command = "python3 src/python/RampUpTime.py"
	r := subprocess.New(command, subprocess.Shell)
	r.Exec()
	dat, err := os.ReadFile("RU_Result.txt")
	if err != nil {
		fmt.Println("File open failed")
	}
	command = "rm RU_Result.txt"
	r = subprocess.New(command, subprocess.Shell)
	r.Exec()
	f1, err := strconv.ParseFloat(string(dat), 32)
	// fmt.Println(f1)
	if err != nil {
		fmt.Println("Conversion of string to float didn't work.")
	}
	return f1

}

// * END OF RAMP-UP TIME * \\

// * START OF BUS FACTOR * \\

// Function to get bus factor metric score
func getBusFactor(url string) float64 {
	make_shortlog_file(url)
	regex, _ := regexp.Compile("[0-9]+") //Regex for parsing count into only integer

	short_log_raw_data, err1 := os.ReadFile(url + "/shortlog.txt")
	if err1 != nil {
		fmt.Println("Did not find shortlog file")
		log.Fatal(err1)
	}

	arr := strings.Split(string(short_log_raw_data), "\n") // parsing shortlog file by lines
	
	len_log := len(arr) - 1
	
	if len_log < 1{
		fmt.Println("No committers for repo " + url)
		delete_shortlog_file(url)
		return 0
	}
	
	var num_bus_committers int
	if len_log < 100{
		num_bus_committers = 1
	}else{
		num_bus_committers = len_log / 100
	}

	total := 0
	total_bus_guys := 0
	var num string

	fmt.Println(len_log)

	for i := 0; i < len_log; i++ {
		num = regex.FindString(arr[i])
		num_int, err2 := strconv.Atoi(num)
		if err2 != nil{
			fmt.Println("Conversion from string to int didn't work (bus factor calc)")
			log.Fatal(err2)
		}
		total += num_int
		if i < num_bus_committers{
			total_bus_guys += num_int
		}
	}
	delete_shortlog_file(url)
	metric := (float64(total) - float64(total_bus_guys)) / float64(total)
	fmt.Println(metric)
	return metric
}


func make_shortlog_file(url string){
	os.Chdir(url)

	cmd := exec.Command("git","shortlog","HEAD","-se", "-n")
	cwd, _ := os.Getwd()

	fmt.Println("dir is " + cwd)

	out,err := cmd.Output()

	if err != nil{
		fmt.Println("Did not find closed issues file from api, invalid url: " + url)
		// log.Fatal(err)	
	}

	os.WriteFile("shortlog.txt", out, 0644)

	os.Chdir("../")
	
	cwd, _ = os.Getwd()
	fmt.Println("dir is " + cwd)

}

func delete_shortlog_file(url string){
	var command string
	command = "rm -f " + url + "/shortlog.txt"
	s := subprocess.New(command, subprocess.Shell)
	s.Exec()

}

// * END OF BUS FACTOR * \\

// * START OF CORRECTNESS * \\

// Function to get correctness metric score
func getCorrectness(url string) float64 {
	// fmt.Println("Getting correctness...")

	runRestApi(url)

	regex, _ := regexp.Compile("\"total_count\": [0-9]+") //Regex for finding count of issues in input file
	num_regex, _ := regexp.Compile("[0-9]+")              //Regex for parsing count into only integer

	//closed issues
	data_closed, err1 := os.ReadFile("./src/issues/closed.txt")
	closed_count := regex.FindString(string(data_closed))
	closed_count = num_regex.FindString(closed_count)
	
	//open issues
	data_open, err := os.ReadFile("./src/issues/open.txt")
	open_count := regex.FindString(string(data_open))
	open_count = num_regex.FindString(open_count)
	if err != nil || err1 != nil{
		fmt.Println("Did not find issues file from api, invalid url: " + url)
		// log.Fatal(err)
		return 0
	}
	// fmt.Println("Open: " + open_count + "\nClosed: " + closed_count)
	
	score := calc_score(open_count, closed_count)
	if math.IsNaN(score) {
		score = 0
	}
	// fmt.Println(score)

	teardownRestApi()

	return score
}

func runRestApi(url string) int {

	index := strings.Index(url, ".com/")
	if index == -1 {
		// fmt.Println("No '.com/' found in the string")
		return 1
	}
	url = url[index+5:]
	
	token := os.Getenv("GITHUB_TOKEN")
	command := "python3 -c 'import os; os.remove(\"src/issues/closed.txt\") if os.path.exists(\"src/issues/closed.txt\") else \"continue\"; os.remove(\"src/issues/open.txt\") if os.path.exists(\"src/issues/open.txt\") else \"continue\"; os.system(\"curl -i -H \\\"Authorization: token "+token+"\\\" https://api.github.com/search/issues?q=repo:"+url+"+type:issue+state:closed >> src/issues/closed.txt\"); os.system(\"curl -i -H \\\"Authorization: token "+token+"\\\" https://api.github.com/search/issues?q=repo:"+url+"+type:issue+state:open >> src/issues/open.txt\");'"

	r := subprocess.New(command, subprocess.Shell)
	r.Exec()
	return 0
}

func teardownRestApi() {
	setup := "import sys; sys.path.append('../'); from src.python import rest_api;"
	cmd := exec.Command("python3", "-c", setup+"rest_api.deleteIssues()")
	_, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
}

func calc_score(s1 string, s2 string) float64 {

	// fmt.Println(s1)
	// fmt.Println(s2)
	f1, err := strconv.ParseFloat(s1, 32)
	if err != nil {
		fmt.Println("Conversion of s1 to string float didn't work.")
	}
	f2, err1 := strconv.ParseFloat(s2, 32)
	if err1 != nil {
		fmt.Println("Conversion of s2 to string float didn't work.")
	}

	f3 := f2 / (f1 + f2)

	return f3
}

// * END OF CORRECTNESS * \\

// * START OF LICENSE COMPATABILITY * \\

// Function to get license compatibility metric score
func getLicenseCompatibility(url string) float64 {
	// fmt.Println("Checking for license... ")

	foundLicense := searchForLicenses("./src/repos/rnd/")

	if foundLicense {
		fmt.Println("[LICENSE FOUND]")
		return 1
	}
	fmt.Println("[LICENSE NOT FOUND]")
	return 0
}

func searchForLicenses(folder string) bool {
	found := false
	//walk the repo looking for the license
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if found {
			return nil
		}
		if info == nil {
			return nil
		}
		if info.IsDir() {
			if len(info.Name()) > 0 && info.Name()[0] == '.' {
				return filepath.SkipDir
			}
		} else {
			found = checkFileForLicense(path)
		}
		return nil
	})

	//catch errors
	if err != nil {
		fmt.Println(err)
	}
	return found
}

func checkFileForLicense(path string) bool {
	license := "LGPL-2.1"
	file, err := os.Open(path)
	if err != nil {
		// fmt.Print("Coudln't open path ")
		fmt.Println(err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, license); idx != -1 {
			fmt.Println("Found license in file:", path)
			return true
		}
	}
	return false
}

// * END OF LICENSE COMPATABILITY * \\

// * START OF REPO CLONING/REMOVING  * \\

func cloneRepo(url string) string {
	fmt.Println("Cloning Repo")
	s := subprocess.New("git clone "+url+" src/repos/rnd", subprocess.Shell)
	if err := s.Exec(); err != nil {
		log.Fatal(err)
		fmt.Println(err)
		return ("ERROR CLONING")
	}
	index := strings.Index(url, ".com/")
	if index == -1 {
		fmt.Println("No '.com/' found in the string")
		return "FAILURE"
	}
	
	url = url[index+5:]
	r, _ := regexp.Compile("/")
	a := r.Split(url, 2)
	return a[1]
}

func clearRepoFolder() {
	s := subprocess.New("rm -rf ", subprocess.Arg("src/repos/*"), subprocess.Shell)
	s.Exec()
}

// * END OF REPO CLONING/REMOVING  * \\

// * START OF STDOUT * \\

func printRepo(next *repo) {
	for {
		if next.URL != "temp" {
			repoOUT(next)
		}
		if next.next == nil {
			break
		}
		next = next.next
	}
}

func repoOUT(r *repo) {
	fmt.Print("{\"URL\":\"", r.URL, "\", \"NET_SCORE\":", r.netScore, ", \"RAMP_UP_SCORE\":", r.rampUpTime, ",\"CORRECTNESS_SCORE\":", r.correctness, ", \"BUS_FACTOR_SCORE\":", r.busFactor, ", \"RESPONSIVE_MAINTAINER_SCORE\":", r.responsiveness, ", \"LICENSE_SCORE\":", r.licenseCompatibility, "}\n")
}

// * END OF STDOUT * \\

// * START OF SORTING * \\

func addRepo(head *repo, curr *repo, temp *repo) *repo {
	head.next = curr
	if curr == nil {
		head.next = temp
	} else {
		if curr.netScore >= temp.netScore {
			curr = addRepo(curr, curr.next, temp)
		} else {
			head.next = temp
			temp.next = curr
		}
	}

	return head
}

// * END OF SORTING * \\