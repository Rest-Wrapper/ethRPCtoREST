package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"regexp"

	"github.com/joho/godotenv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/gofiber/fiber/v2"
)

var RPC_URL string
var client *ethclient.Client

func main() {
	app := fiber.New()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading environment variables file")
	}

	RPC_URL = os.Getenv("RPC_URL")
	log.Println(RPC_URL)

	client, err = ethclient.Dial(RPC_URL)
	if err != nil {
		log.Fatal(err)
	}

	app.Get("/eth/block/:identifier", getBlockByIdentifier)

	log.Fatal(app.Listen(":3000"))
}

// Handler
func getBlockByIdentifier(c *fiber.Ctx) error {
	identifier := c.Params("identifier")

	// Check if identifier is a block hash or block number
	// A Block hash is 32 bytes long and hence 64 characters long without 0x prefix
	// A block number will always consist of a non-zero character after 0x, except for "0x0".

	// Use regex to check if identifier is a block hash or block number
	blockHashRegex := regexp.MustCompile(`^0x[0-9a-f]{64}$`)

	// A block number also allows default block identifiers such as "earliest", "latest" and "pending"
	// TODO: A block number can also be a decimal number without 0x prefix (my proposal)
	// A block number can also be a hex number with 0x prefix

	// Regex to allow for default block identifiers
	blockNumberRegex := regexp.MustCompile(`^0x([1-9a-f]+[0-9a-f]*|0)$|^earliest$|^latest$|^pending$|^safe$|^finalized$`)

	// Check if identifier is a block hash or block number
	if blockHashRegex.MatchString(identifier) {
		log.Println("Block hash")
		blockInfo := getBlockByHash(c, identifier)
		return blockInfo
	} else if blockNumberRegex.MatchString(identifier) {
		log.Println("Block number")
		blockInfo := getBlockByNumber(c, identifier)
		return blockInfo
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid identifier",
		})
	}
}

// getBlockByHash retrieves block information by hash and returns it as JSON.
func getBlockByHash(c *fiber.Ctx, hash string) error {
	blockHash := common.HexToHash(hash)
	log.Println(blockHash)
	block, err := client.HeaderByHash(context.Background(), blockHash)
	if err != nil {
		log.Fatal("Error fetching block info:", err)
	}
	return c.JSON(block)
}

// getBlockByNumber retrieves block information by block number or default block parameters and returns it as JSON.
func getBlockByNumber(c *fiber.Ctx, numberOrDefaultParameters string) error {
	//TODO: Add support for default block parameters
	number := numberOrDefaultParameters
	number = number[2:] // Remove 0x prefix
	log.Println(number)

	blockNumber, success := new(big.Int).SetString(number, 16)
	if !success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid number",
		})
	}
	log.Println(blockNumber)
	blockInfo, err := client.HeaderByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal("Error fetching block info:", err)
	}
	return c.JSON(blockInfo)
}
