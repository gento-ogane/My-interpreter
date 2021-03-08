package main

import (
	"fmt"
	"monkey/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Welcome %s! This is the げんまる programming language!\n\n",
		user.Username)
	fmt.Printf(`
          #  # #       #                  #                         
  #       #  # #       #                  #           #########     
  #       #           ##                  #                 ##      
 ##       #           #             ############           ##       
 #   ##########      ##                   #               ##        
 #        #          #                    #              ##         
 #        #          #              ############        #######     
 #        #         ## ###                #            ##     ##    
 #        #         # #   #               #           ##       ##   
 #        #         ##    #               #          ##         #   
 # #      #        ##     #          ######             ###     #   
 ##      ##        #      #    #    #     ###          #  ##    #   
  #      #         #      #   ##    #     # ###        #   #   ##   
        ##        ##      #  ##     #    ##   ##       ##  #  ##    
       ##         #        ###       #####              #######     
                                                                    
`)
	repl.Start(os.Stdin, os.Stdout)
}
