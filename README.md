<h1>Solution to 'Mandatory Hand-In 4'</h1>

Repository for the solution to 'Mandatory Hand-In 4 - Distributed Mutual Exclusion, Distributed Systems, BSc (Autumn 2023)'. Made by Group 17. 

<h2>HOW TO RUN:</h2>
In order to run this program, do as follows:
<ol>
  <li>Change directory into the <b>homework4</b>-folder</li>
  <li>Run the following command: <i>go run node.go</i></li>
  <li>Open as many new terminals as you'd like, and redo the previous steps</li>
  <li>Most likely the program will create a new file, <b>criticalsection.txt</b>, which the different nodes then will try to access at different times </li>
</ol> 

<h3>Alternatively: run as executable binaries</h3>
You can also choose to run the peer nodes as executable binaries - like a user would in real life. 
In order to do so, we recommend doing the following:
<ol>
  <li>Make sure that your <b>GOPATH</b> environment variable is set - the generated binaries will go here</li>
  <li>Change directory into the <b>homework4</b>-folder</li>
  <li>In this directory, open a new terminal and run the following command: <i>go install</i></li>
  <li>Now an executable binary of the peer node will be available at your <b>GOPATH</b> (likely inside a 'bin' file, with the executable name <b>homework4</b>) </li>
</ol> 
