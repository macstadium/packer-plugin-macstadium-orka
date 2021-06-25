## Development / Bugs / Support

If you want to help contribute to this plugin or find a bug, please [file an issue] on Github, and/or submit me a PR.  This is Open Source, so don't expect me to fix bugs immediately, but I'll try my best to reasonably support this plugin.  Contributors are always welcome though.

To get a development environment up will need a recent golang installed and setup.  Then with a single make command below command it will build, install, and try to run the example at [examples/orka.pkr.hcl](./examples/orka.pkr.hcl).  You may need to edit this file to have the source VM config that you have locally, as it is hardcoded to my environment's image `macos-catalina-10-15-5` at the moment.

```bash
make fresh
```

or  

```bash
make rebuild
```

## Todo (in no particular order)

These are a list of things that are pending to accomplish within' this repo.  Contributors welcome, I might do some of these also, eventually.

 * Add more tests 
 * Add image management features into the builder instead of having to manage out-of-band with orka cli (eg: delete image)
 * Consider implementing for this plugin to automatically create a VM Config (before launching a VM, and possibly after tied to the image just created)
 * Add the ability for this plugin to be able to scan configs and/or images available and automatically use the first one it finds, possibly creating a new config specifically for the image?  
 * Improve the JSON parsing code.  See: line 137-176 of [builder/orka/step_orka_create.go](./builder/orka/step_orka_create.go) and look at the function `ExtractIPHost` in that file as well.  Could be much improved and hopefully simplified.  Contributors welcome!!!
 * Clean up / improve code / catch more sharp edges and edge-cases, deal with any issues filed on Github.
 * One day, get this merged upstream into Packer as an official packer plugin.  Additional info [here](https://www.packer.io/docs/plugins/packer-integration-program)



[//]: <> (Ignore, below here are links for ease-of-use above)
[file an issue]: https://github.com/macstadium/packer-plugin-macstadium-orka/issues