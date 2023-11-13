<h1>Solution to 'Mandatory Hand-In 3'</h1>

Repository for the solution to 'Mandatory Hand-In 3 - Chitty Chat, Distributed Systems, BSc (Autumn 2023)'. Made by Group 17. 

<h2>HOW TO RUN:</h2>
There are two important folders associated with this repository: the <b>client</b>-folder, and the <b>server</b>-folder.
In order to run this program, do as follows:
<ol>
  <li>Change directory into the <b>server</b>-folder</li>
  <li>Run the following command: <i>go run server.go</i></li>
  <li>After, change directory into the <b>client</b>-folder</li>
  <li>Now you can run the following command any number of times: <i>go run client.go</i></li>
  <li>Everytime you run <i>'go run client.go'</i>, a new client application will start </li>
</ol> 

<h3>Alternatively: run Chitty Chat as executable binaries</h3>
You can also choose to run the server and client as executable binaries - like a user would in real life. 
In order to do so, we recommend doing the following:
<ol>
  <li>Make sure that your <b>GOPATH</b> environment variable is set - the generated binaries will go here</li>
  <li>Change directory into the <b>server</b>-folder</li>
  <li>Run the following command: <i>go install</i></li>
  <li>After, change directory into the <b>client</b>-folder</li>
  <li>Run the following command: <i>go install</i></li>
  <li>Now an executable binary of both the client and server will be available at your <b>GOPATH</b> (likely inside a 'bin' file) </li>
</ol> 