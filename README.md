# <p align="center"><img src="https://github.com/SunProxy/sun/blob/master/SunProxy.png"/></p>

# Downloads
<a align="center">[Pipelines](https://app.circleci.com/pipelines/github/SunProxy/sun) </a><br>
Here you can find all the build please select the latest and click Artifacts.

## Explanation
A normal connection is made out to be like this. <br >

Client | Direction | Server
------------ | ------------- | -------------
Drops item |  -> | Processes request
Spawns Item on the ground | <- | Sends back a inventory transaction and AddActor packet

That is known as Peer To Peer. Now we get into the more juicy graphs.

Client | Direction | Proxy | Direction | Server
------------ | ------------- | ------------- | ------------- | -------------
Drops item | -> | Forwards request to the server | -> | Processes request
Spawns Item on the ground | <- | Forwards request to the Client | <- | Sends back a inventory transaction and AddActor packet

This is called man in the middle proxying. <br>
It allows for hacking proxies and custom packets / behavior to work no matter what server you are on. <br>
SunProxy makes use of this for our Custom Transfer and Messaging system!

# Discussion
<a align="center">[Discord](https://discord.gg/g4SJUffja3) </a><br>
# License
```
      ___           ___           ___
     /  /\         /__/\         /__/\
    /  /:/_        \  \:\        \  \:\
   /  /:/ /\        \  \:\        \  \:\
  /  /:/ /::\   ___  \  \:\   _____\__\:\
 /__/:/ /:/\:\ /__/\  \__\:\ /__/::::::::\
 \  \:\/:/~/:/ \  \:\ /  /:/ \  \:\~~\~~\/
  \  \::/ /:/   \  \:\  /:/   \  \:\  ~~~
   \__\/ /:/     \  \:\/:/     \  \:\
     /__/:/       \  \::/       \  \:\
     \__\/         \__\/         \__\/

MIT License

Copyright (c) 2020 Jviguy

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
