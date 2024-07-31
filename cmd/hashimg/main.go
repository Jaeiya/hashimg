package main

import (
	"fmt"
	"os"

	"github.com/jaeiya/go-template/internal/lib"
)

/*
1. Hash all images that do no contain the special prefix (0x@)

  - All images with this prefix have their hash as their file name

  - All images are hashed using the sha256 algorithm

  - Hash should be truncated to a specific size

    2. Load all image names with the special prefix (0x@) and add them to
    their own list.

  - At this point we should have two lists of hashes

  - A list of hashes to images that are NOT processed       (NewImageHashes)

  - A list of hashes to images that are already processed   (OldImageHashes)

  - Old hashes should be added with the prefix stripped

2a. Images that are loaded into the NewImageHashes list, should be checked

	for conflicts. If an incoming image, which is bound for the NewImageHashes
	list, conflicts with an existing hash within the NewImageHashes list,
	delete the incoming image.

2b. Images that are loaded into the OldImageHashes list, should be assumed

		to all be unique and do NOT require comparison for conflicts.

	 3. Compare the two lists, such that any hashes from the NewImageHashes list which
	    are found within the OldImageHashes list, should be deleted from the
	    NewImageHashes list.
	    - After this process, both lists should contain unique hashes that
	    cannot be found within the other list.

	 4. Rename all images from the NewImageHashes list, to their respective hash
	    value, along with the special prefix (0x@).
	    - Example: 0x@3aab051a5ef6ca5b638c0f9 (truncated to custom value)
	    - The renaming process should be executed with as many go routines
	    as is reasonable, to be as fast as possible.

4a. If applicable, use a thread pool for not only the renaming of the

	image files, but also the hashing process for NewImageHashes.
	- It will be a lot faster to process new images if we hash them
	  in parallel.
	- The only concern here, is how much memory this can take
	  because each image needs to be fully loaded, before it can
	  be hashed.
*/
func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fmt.Println("\nProcessing: ", wd)
	iMap, err := lib.MapImages(wd)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nLoaded:", len(iMap), "images")
	stats, err := lib.ProcessImages(wd, 32, iMap)
	if err != nil {
		panic(err)
	}

	fmt.Println("Cached:", stats.New)
	fmt.Println(" Dupes:", stats.Dup)
}
