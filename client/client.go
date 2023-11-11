package client

// CS 161 Project 2

// You MUST NOT change these default imports. ANY additional imports
// may break the autograder!

import (
	"encoding/json"

	userlib "github.com/cs161-staff/project2-userlib"
	"github.com/google/uuid"

	// hex.EncodeToString(...) is useful for converting []byte to string

	// Useful for string manipulation
	//"strings"

	// Useful for formatting strings (e.g. `fmt.Sprintf`).
	"fmt"

	// Useful for creating new error messages to return using errors.New("...")
	"errors"

	// Optional.
	_ "strconv"
)

// This serves two purposes: it shows you a few useful primitives,
// and suppresses warnings for imports not being used. It can be
// safely deleted!
func someUsefulThings() {

	// Creates a random UUID.
	randomUUID := uuid.New()

	// Prints the UUID as a string. %v prints the value in a default format.
	// See https://pkg.go.dev/fmt#hdr-Printing for all Golang format string flags.
	userlib.DebugMsg("Random UUID: %v", randomUUID.String())

	// Creates a UUID deterministically, from a sequence of bytes.
	hash := userlib.Hash([]byte("user-structs/alice"))
	deterministicUUID, err := uuid.FromBytes(hash[:16])
	if err != nil {
		// Normally, we would `return err` here. But, since this function doesn't return anything,
		// we can just panic to terminate execution. ALWAYS, ALWAYS, ALWAYS check for errors! Your
		// code should have hundreds of "if err != nil { return err }" statements by the end of this
		// project. You probably want to avoid using panic statements in your own code.
		panic(errors.New("An error occurred while generating a UUID: " + err.Error()))
	}
	userlib.DebugMsg("Deterministic UUID: %v", deterministicUUID.String())

	// Declares a Course struct type, creates an instance of it, and marshals it into JSON.
	type Course struct {
		name      string
		professor []byte
	}

	course := Course{"CS 161", []byte("Nicholas Weaver")}
	courseBytes, err := json.Marshal(course)
	if err != nil {
		panic(err)
	}

	userlib.DebugMsg("Struct: %v", course)
	userlib.DebugMsg("JSON Data: %v", courseBytes)

	// Generate a random private/public keypair.
	// The "_" indicates that we don't check for the error case here.
	var pk userlib.PKEEncKey
	var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("PKE Key Pair: (%v, %v)", pk, sk)

	// Here's an example of how to use HBKDF to generate a new key from an input key.
	// Tip: generate a new key everywhere you possibly can! It's easier to generate new keys on the fly
	// instead of trying to think about all of the ways a key reuse attack could be performed. It's also easier to
	// store one key and derive multiple keys from that one key, rather than
	originalKey := userlib.RandomBytes(16)
	derivedKey, err := userlib.HashKDF(originalKey, []byte("mac-key"))
	if err != nil {
		panic(err)
	}
	userlib.DebugMsg("Original Key: %v", originalKey)
	userlib.DebugMsg("Derived Key: %v", derivedKey)

	// A couple of tips on converting between string and []byte:
	// To convert from string to []byte, use []byte("some-string-here")
	// To convert from []byte to string for debugging, use fmt.Sprintf("hello world: %s", some_byte_arr).
	// To convert from []byte to string for use in a hashmap, use hex.EncodeToString(some_byte_arr).
	// When frequently converting between []byte and string, just marshal and unmarshal the data.
	//
	// Read more: https://go.dev/blog/strings

	// Here's an example of string interpolation!
	_ = fmt.Sprintf("%s_%d", "file", 1)
}

// This is the type definition for the User struct.
// A Go struct is like a Python or Java class - it can have attributes
// (e.g. like the Username attribute) and methods (e.g. like the StoreFile method below).
type User struct {
	Username  string
	Password  []byte
	Salt      []byte
	fly_key_1 []byte
	fly_key_2 []byte

	pKEDecKey          userlib.PKEDecKey
	EncryptedPKEDecKey []byte

	dSSignKey          userlib.DSSignKey
	EncryptedDSSignKey []byte

	FileNameSpaceUUID uuid.UUID
	fileNameSpace     FileNameSpace //map from file name to file access node UUID

	HashCheck []byte
}

type FileNameSpace struct {
	fileNameSpace          map[string]uuid.UUID //map from file name to file access node UUID
	EncryptedFileNameSpace []byte
	HashCheck              []byte
}

type File struct {
	// first
	ArthorizedUsers     map[string]string //maps uuid of accepter to uuid of inviter
	ArthorizedAccess    map[string]uuid.UUID //maps uuid of accepter to its file access node
	HeadContentNodeUUID uuid.UUID
	TailContentNodeUUID uuid.UUID

	hashCheck          []byte
	EncryptedHashCheck []byte
}

type FileAccessNode struct {
	SelfUUID uuid.UUID

	fileUUID          uuid.UUID
	EncryptedFileUUID []byte

	symmetricEncryptionKey          []byte
	EncryptedSymmetricEncryptionKey []byte

	NewSymmetricEncryptionKeyProvider string
	NewEncryptedSymmetricEncryptionKey []byte
	DSSignNewSymmetricEncryptionKey []byte

	hashCheck          []byte
	EncryptedHashCheck []byte
}

type ContentNode struct {
	SelfUUID uuid.UUID
	NextUUID uuid.UUID

	content          []byte
	EncryptedContent []byte

	hashCheck          []byte
	EncryptedHashCheck []byte
}

// NOTE: The following methods have toy (insecure!) implementations.
func InitUser(username string, password string) (userdataptr *User, err error) {
	var userdata User

	// if username provided is empty return error
	if len(username) == 0 {
		return nil, errors.New("username cannot be empty")
	}
	// if exists same username return error
	userUUID, err := uuid.FromBytes(userlib.Hash([]byte(username))[:16])
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	userNameExisting, ok := userlib.DatastoreGet(userUUID)
	// skip declared and not used
	// _ = userNameExisting

	if ok || !(len(userNameExisting) == 0) {
		return nil, errors.New("an error occurred")
	}

	// initUser
	salt := userlib.RandomBytes(32)
	// not sure the len
	keyLen := uint32(32)

	userdata.Username = username
	userdata.Salt = salt
	userdata.Password = userlib.Argon2Key([]byte(password), salt, keyLen)

	// fly key
	flyKey_1, err := userlib.HashKDF([]byte(userlib.Hash(([]byte(username + password + "u then p"))))[:16], []byte("fly key"))
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	flyKey_2, err := userlib.HashKDF([]byte(userlib.Hash(([]byte(password + username + "p then u"))))[:16], []byte("fly key"))
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	userdata.fly_key_1 = flyKey_1
	userdata.fly_key_2 = flyKey_2

	// Generate public key and private key
	PKEEncKey, PKEDecKey, err := userlib.PKEKeyGen()
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	userlib.KeystoreSet(username+"PKEEncKey", PKEEncKey)
	userdata.pKEDecKey = PKEDecKey
	PKEDecKeyBytes, err := json.Marshal(PKEDecKey)
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	userdata.EncryptedPKEDecKey = userlib.SymEnc(flyKey_1[0:16], salt[0:16], PKEDecKeyBytes)

	// Generate Digital Signature
	DSSignKey, DSVerifyKey, err := userlib.DSKeyGen()
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	userlib.KeystoreSet(username+"DSVerifyKey", DSVerifyKey)
	userdata.dSSignKey = DSSignKey
	DSSignKeyBytes, err := json.Marshal(DSSignKey)
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	userdata.EncryptedDSSignKey = userlib.SymEnc(flyKey_1[16:32], salt[0:16], DSSignKeyBytes)

	// ini fileNamespace
	// userdata.fileNameSpace = make(map[string]uuid.UUID)
	userdata.FileNameSpaceUUID, err = userdata.initFileNameSpace()
	if err != nil {
		return nil, errors.New("an error occurred")
	}

	FileNameSpaceUUIDByte, err := json.Marshal(userdata.FileNameSpaceUUID)
	if err != nil {
		return nil, errors.New("an error occurred")
	}

	//Hash the above content and make sure the struct isn't modified by malicious action
	//HashByte := append([]byte(userdata.Username), userdata.Password, userdata.Salt, userdata.EncryptedPKEDecKey, userdata.EncryptedDSSIgnKey, userdata.EncryptedFileNameSpace)
	HashByte := []byte(userdata.Username)
	HashByte = append(HashByte, userdata.Password...)
	HashByte = append(HashByte, userdata.Salt...)
	HashByte = append(HashByte, userdata.EncryptedPKEDecKey...)
	HashByte = append(HashByte, userdata.EncryptedDSSignKey...)
	HashByte = append(HashByte, FileNameSpaceUUIDByte...)
	HashByte = append(HashByte, flyKey_1[48:63]...)
	userdata.HashCheck, err = userlib.HashKDF(userlib.Hash(HashByte)[:16], []byte("Check Not Modified"))

	userdataBytes, err := json.Marshal(userdata)
	userlib.DatastoreSet(userUUID, userdataBytes)
	return &userdata, nil
}

func GetUser(username string, password string) (userdataptr *User, err error) {
	// check user exit or not
	userUUID, err := uuid.FromBytes(userlib.Hash([]byte(username))[:16])
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	user, ok := userlib.DatastoreGet(userUUID)
	if !ok {
		return nil, errors.New("an error occurred")
	}

	// user exit, get user
	var userdata User
	json.Unmarshal(user, &userdata)
	keyLen := uint32(32)
	passwordMatch := userlib.HMACEqual(userlib.Argon2Key([]byte(password), userdata.Salt, keyLen), userdata.Password)
	// password doesn't match return error
	if !passwordMatch {
		return nil, errors.New("an error occurred")
	}

	//re-derive fly-key
	flyKey_1, err := userlib.HashKDF([]byte(userlib.Hash(([]byte(username + password + "u then p"))))[:16], []byte("fly key"))
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	flyKey_2, err := userlib.HashKDF([]byte(userlib.Hash(([]byte(password + username + "p then u"))))[:16], []byte("fly key"))
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	userdata.fly_key_1 = flyKey_1
	userdata.fly_key_2 = flyKey_2

	FileNameSpaceUUIDByte, err := json.Marshal(userdata.FileNameSpaceUUID)
	if err != nil {
		return nil, errors.New("an error occurred")
	}

	//check if struct has been modified by malicious action
	HashByte := []byte(userdata.Username)
	HashByte = append(HashByte, userdata.Password...)
	HashByte = append(HashByte, userdata.Salt...)
	HashByte = append(HashByte, userdata.EncryptedPKEDecKey...)
	HashByte = append(HashByte, userdata.EncryptedDSSignKey...)
	HashByte = append(HashByte, FileNameSpaceUUIDByte...)
	HashByte = append(HashByte, flyKey_1[48:63]...)
	localHashCheck, err := userlib.HashKDF(userlib.Hash(HashByte)[:16], []byte("Check Not Modified"))
	if err != nil {
		return nil, errors.New("an error occurred")
	}
	HashCheckMatch := userlib.HMACEqual(localHashCheck, userdata.HashCheck)
	if !HashCheckMatch {
		return nil, errors.New("an error occurred")
	}

	//re-derive other keys
	PKEDecKeyBytes := userlib.SymDec(flyKey_1[0:16], userdata.EncryptedPKEDecKey)
	err = json.Unmarshal(PKEDecKeyBytes, &userdata.pKEDecKey)
	if err != nil {
		return nil, errors.New("an error occurred")
	}

	DSSignKeyBytes := userlib.SymDec(flyKey_1[16:32], userdata.EncryptedDSSignKey)
	err = json.Unmarshal(DSSignKeyBytes, &userdata.dSSignKey)
	if err != nil {
		return nil, errors.New("an error occurred")
	}

	//get file name space
	/*fileNameSpaceByte := userlib.SymDec(flyKey_1[32:48], userdata.EncryptedFileNameSpace)
	err = json.Unmarshal(fileNameSpaceByte, &userdata.fileNameSpace)
	if err != nil {
		return nil, errors.New("an error occurred")
	}*/

	userdata.fileNameSpace, err = userdata.getFileNameSpace()
	if err != nil {
		return nil, errors.New("an error occurred")
	}

	userdataptr = &userdata
	return userdataptr, nil
}

func (userdata *User) initFileNameSpace() (fileNameSpaceUUID uuid.UUID, err error) {
	var newFileNameSpace FileNameSpace
	newFileNameSpace.fileNameSpace = make(map[string]uuid.UUID)
	fileNameSpaceByte, err := json.Marshal(newFileNameSpace.fileNameSpace)
	if err != nil {
		return uuid.Nil, errors.New("an error occurred")
	}
	newFileNameSpace.EncryptedFileNameSpace = userlib.SymEnc(userdata.fly_key_1[32:48], userdata.Salt[0:16], fileNameSpaceByte)

	HashByte := append(newFileNameSpace.EncryptedFileNameSpace, userdata.fly_key_2[0:16]...)
	newFileNameSpace.HashCheck, err = userlib.HashKDF(userlib.Hash(HashByte)[:16], []byte("Check Not Modified"))
	if err != nil {
		return uuid.Nil, errors.New("an error occurred")
	}

	fileNameSpaceUUID = uuid.New()

	newFileNameSpaceByte, err := json.Marshal(newFileNameSpace)
	if err != nil {
		return uuid.Nil, errors.New("an error occurred")
	}

	userlib.DatastoreSet(fileNameSpaceUUID, newFileNameSpaceByte)
	return fileNameSpaceUUID, nil
}

func (userdata *User) getFileNameSpace() (node FileNameSpace, err error) {
	var fileNameSpace FileNameSpace
	fileNameSpaceByte, ok := userlib.DatastoreGet(userdata.FileNameSpaceUUID)
	if !ok {
		return fileNameSpace, errors.New("getFileNameSpace an error occurred 1")
	}
	err = json.Unmarshal(fileNameSpaceByte, &fileNameSpace)
	if err != nil {
		return fileNameSpace, errors.New("an error occurred")
	}

	HashByte := append(fileNameSpace.EncryptedFileNameSpace, userdata.fly_key_2[0:16]...)
	localHashCheck, err := userlib.HashKDF(userlib.Hash(HashByte)[:16], []byte("Check Not Modified"))
	ok = userlib.HMACEqual(localHashCheck, fileNameSpace.HashCheck)
	if !ok {
		return fileNameSpace, errors.New("file name space modified!")
	}

	fileNameSpacefileNameSpaceByte := userlib.SymDec(userdata.fly_key_1[32:48], fileNameSpace.EncryptedFileNameSpace)
	err = json.Unmarshal(fileNameSpacefileNameSpaceByte, &fileNameSpace.fileNameSpace)
	if err != nil {
		return fileNameSpace, errors.New("an error occurred")
	}
	return fileNameSpace, nil
}

// get File access node by its uuid
func (userdata *User) getFileAccessNode(fileAccessNodeUUID uuid.UUID) (fileAccessNode FileAccessNode, err error) {
	fileAccessNodeByte, ok := userlib.DatastoreGet(fileAccessNodeUUID)

	if !ok {
		return fileAccessNode, errors.New("getFileAccessNode an error occurred 1")
	}
	err = json.Unmarshal(fileAccessNodeByte, &fileAccessNode)

	if err != nil {
		return fileAccessNode, errors.New("getFileAccessNode an error occurred 2")
	}

	//check that fileAccessNode isn't modified by malicious action
	selfUUIDByte, err := json.Marshal(fileAccessNode.SelfUUID)
	if err != nil {
		return fileAccessNode, errors.New("getFileAccessNode an error occurred 3")
	}

	FileAccessNodeHashByte := selfUUIDByte
	FileAccessNodeHashByte = append(FileAccessNodeHashByte, fileAccessNode.EncryptedFileUUID...)
	FileAccessNodeHashByte = append(FileAccessNodeHashByte, fileAccessNode.EncryptedSymmetricEncryptionKey...)
	FileAccessNodeHashByte = append(FileAccessNodeHashByte, userdata.fly_key_2[16:32]...)
	localHashCheck := userlib.Hash(FileAccessNodeHashByte)

	fileAccessNode.hashCheck, err = userlib.PKEDec(userdata.pKEDecKey, fileAccessNode.EncryptedHashCheck)
	if err != nil {
		return fileAccessNode, errors.New("getFileAccessNode an error occurred 4")
	}

	if !userlib.HMACEqual(localHashCheck, fileAccessNode.hashCheck) {
		return fileAccessNode, errors.New("getFileAccessNode an error occurred 5")
	}

	//decrypt file access node
	fileUUIDByte, err := userlib.PKEDec(userdata.pKEDecKey, fileAccessNode.EncryptedFileUUID)
	if err != nil {
		return fileAccessNode, errors.New("getFileAccessNode an error occurred 6")
	}
	err = json.Unmarshal(fileUUIDByte, &fileAccessNode.fileUUID)
	if err != nil {
		return fileAccessNode, errors.New("getFileAccessNode an error occurred 7")
	}

	fileAccessNode.symmetricEncryptionKey, err = userlib.PKEDec(userdata.pKEDecKey, fileAccessNode.EncryptedSymmetricEncryptionKey)
	if err != nil {
		return fileAccessNode, errors.New("getFileAccessNode an error occurred 8")
	}

	//check if a new key for the file is provided, must check if it is authorized to access the file
	if (fileAccessNode.NewEncryptedSymmetricEncryptionKey != nil) {

		//check if provider is authorized to access the file
		var file File

		fileByte, ok := userlib.DatastoreGet(fileAccessNode.fileUUID)

		if !ok {
			return fileAccessNode, errors.New("getFile err 1")
		}

		err = json.Unmarshal(fileByte, &file)
		if err != nil {
			return fileAccessNode, err
		}

		if _, ok := file.ArthorizedUsers[fileAccessNode.NewSymmetricEncryptionKeyProvider]; !ok {
			return fileAccessNode, err
		}

		//verify the digital signature
		providerDSVerifyKey, ok := userlib.KeystoreGet(fileAccessNode.NewSymmetricEncryptionKeyProvider+"DSVerifyKey")
		if !ok {
			return fileAccessNode, errors.New("getFile err 1")
		}

		err = userlib.DSVerify(providerDSVerifyKey, fileAccessNode.NewEncryptedSymmetricEncryptionKey, fileAccessNode.DSSignNewSymmetricEncryptionKey)
		if err != nil {
			return fileAccessNode, err
		}

		//if all verification is successful, set new key to key
		fileAccessNode.EncryptedSymmetricEncryptionKey = fileAccessNode.NewEncryptedSymmetricEncryptionKey
		fileAccessNode.symmetricEncryptionKey, err = userlib.PKEDec(userdata.pKEDecKey, fileAccessNode.EncryptedSymmetricEncryptionKey)
		if err != nil {
			return fileAccessNode, errors.New("getFileAccessNode an error occurred 8")
		}

		//clear unnecessary fields
		fileAccessNode.NewSymmetricEncryptionKeyProvider = ""
		fileAccessNode.NewEncryptedSymmetricEncryptionKey = nil
		fileAccessNode.DSSignNewSymmetricEncryptionKey = nil

		//recompute hash and save the node
		userPKEEncryptKey, ok := userlib.KeystoreGet(userdata.Username + "PKEEncKey")
		if !ok {
			return fileAccessNode, errors.New("an error occurred")
		}
		
		FileAccessNodeHashByte := selfUUIDByte
		FileAccessNodeHashByte = append(FileAccessNodeHashByte, fileAccessNode.EncryptedFileUUID...)
		FileAccessNodeHashByte = append(FileAccessNodeHashByte, fileAccessNode.EncryptedSymmetricEncryptionKey...)
		FileAccessNodeHashByte = append(FileAccessNodeHashByte, userdata.fly_key_2[16:32]...)
		fileAccessNode.hashCheck = userlib.Hash(FileAccessNodeHashByte)
		fileAccessNode.EncryptedHashCheck, err = userlib.PKEEnc(userPKEEncryptKey, fileAccessNode.hashCheck)
		if err != nil {
			return fileAccessNode, err
		}

		fileAccessNodeByte, err := json.Marshal(fileAccessNode)
		if err != nil {
			return fileAccessNode, err
		}

		userlib.DatastoreSet(fileAccessNode.SelfUUID, fileAccessNodeByte)
	}

	return fileAccessNode, nil
}

// create new file accessnode for user, store it before returning
func (userdata *User) newFileAccessNode(username string, symmetricEncryptionKey []byte, initFileUUID uuid.UUID) (fileAccessNode FileAccessNode, err error) {
	//userPKEEncryptKey, ok := userlib.KeystoreGet(userdata.Username + "PKEEncKey")
	userPKEEncryptKey, ok := userlib.KeystoreGet(username + "PKEEncKey")
	if !ok {
		return fileAccessNode, errors.New("an error occurred")
	}

	//initialize new file access node
	fileAccessNode.SelfUUID = uuid.New()
	fileAccessNode.fileUUID = initFileUUID
	selfUUIDByte, err := json.Marshal(fileAccessNode.SelfUUID)
	if err != nil {
		return fileAccessNode, err
	}
	fileUUIDByte, err := json.Marshal(fileAccessNode.fileUUID)
	if err != nil {
		return fileAccessNode, err
	}
	fileAccessNode.EncryptedFileUUID, err = userlib.PKEEnc(userPKEEncryptKey, fileUUIDByte)
	if err != nil {
		return fileAccessNode, err
	}

	//fileAccessNode.symmetricEncryptionKey = userlib.RandomBytes(16)
	fileAccessNode.symmetricEncryptionKey = symmetricEncryptionKey
	fileAccessNode.EncryptedSymmetricEncryptionKey, err = userlib.PKEEnc(userPKEEncryptKey, fileAccessNode.symmetricEncryptionKey)
	if err != nil {
		return fileAccessNode, err
	}

	//if the fileaccessnode is for the user it self, set up the hash with fly key, else let the accepter set it up
	if (userdata.Username == username) {
		FileAccessNodeHashByte := selfUUIDByte
		FileAccessNodeHashByte = append(FileAccessNodeHashByte, fileAccessNode.EncryptedFileUUID...)
		FileAccessNodeHashByte = append(FileAccessNodeHashByte, fileAccessNode.EncryptedSymmetricEncryptionKey...)
		FileAccessNodeHashByte = append(FileAccessNodeHashByte, userdata.fly_key_2[16:32]...)
		fileAccessNode.hashCheck = userlib.Hash(FileAccessNodeHashByte)
		fileAccessNode.EncryptedHashCheck, err = userlib.PKEEnc(userPKEEncryptKey, fileAccessNode.hashCheck)
		if err != nil {
			return fileAccessNode, err
		}
	}

	//set newsymmetricencryption key  and its encryption to Nil, since we are not using theses field yet
	fileAccessNode.NewSymmetricEncryptionKeyProvider = ""
	fileAccessNode.NewEncryptedSymmetricEncryptionKey = nil
	fileAccessNode.DSSignNewSymmetricEncryptionKey = nil

	//store new file access node
	fileAccessNodeByte, err := json.Marshal(fileAccessNode)
	if err != nil {
		return fileAccessNode, err
	}

	userlib.DatastoreSet(fileAccessNode.SelfUUID, fileAccessNodeByte)
	return fileAccessNode, nil
}

// get File according to file access node
func (userdata *User) getFile(fileAccessNode FileAccessNode) (file File, err error) {
	//get file according to fileUUID of file access_node

	fileByte, ok := userlib.DatastoreGet(fileAccessNode.fileUUID)

	if !ok {
		return file, errors.New("getFile err 1")
	}

	err = json.Unmarshal(fileByte, &file)
	if err != nil {
		return file, err
	}

	//chcek that file isn't modified by malicious actions
	AuthorizedUsersByte, err := json.Marshal(file.ArthorizedUsers)
	if err != nil {
		return file, err
	}

	AuthorizedAccesssByte, err := json.Marshal(file.ArthorizedAccess)
	if err != nil {
		return file, err
	}

	headContentNodeUUIDByte, err := json.Marshal(file.HeadContentNodeUUID)
	if err != nil {
		return file, err
	}

	tailContentNodeUUIDByte, err := json.Marshal(file.TailContentNodeUUID)
	if err != nil {
		return file, err
	}

	fileHashByte := AuthorizedUsersByte
	fileHashByte = append(fileHashByte, AuthorizedAccesssByte...)
	fileHashByte = append(fileHashByte, headContentNodeUUIDByte...)
	fileHashByte = append(fileHashByte, tailContentNodeUUIDByte...)

	localHashCheck := userlib.Hash(fileHashByte)
	file.hashCheck = userlib.SymDec(fileAccessNode.symmetricEncryptionKey, file.EncryptedHashCheck)
	ok = userlib.HMACEqual(localHashCheck, file.hashCheck)

	if !ok {
		return file, errors.New("getFile err 2")
	}

	//let file verify user's digital signature
	_, ok = file.ArthorizedUsers[userdata.Username]
	if !ok {
		return file, errors.New("getFile err 3")
	}
	userDSVerifyKey, ok := userlib.KeystoreGet(userdata.Username + "DSVerifyKey")
	if !ok {
		return file, errors.New("getFile err 4")
	}

	dummyMessageByte := userlib.RandomBytes(16)
	dummySign, err := userlib.DSSign(userdata.dSSignKey, dummyMessageByte)
	if err != nil {
		return file, err
	}
	err = userlib.DSVerify(userDSVerifyKey, dummyMessageByte, dummySign)
	if err != nil {
		return file, err
	}

	return file, nil
}

func (userdata *User) updateFile(fileAccessNode FileAccessNode, file File) (err error) {
	// recompute hash check
	AuthorizedUsersByte, err := json.Marshal(file.ArthorizedUsers)
	if err != nil {
		return err
	}
	AuthorizedAccesssByte, err := json.Marshal(file.ArthorizedAccess)
	if err != nil {
		return err
	}
	headContentNodeUUIDByte, err := json.Marshal(file.HeadContentNodeUUID)
	if err != nil {
		return err
	}
	tailContentNodeUUIDByte, err := json.Marshal(file.TailContentNodeUUID)
	if err != nil {
		return err
	}

	fileHashByte := AuthorizedUsersByte
	fileHashByte = append(fileHashByte, AuthorizedAccesssByte...)
	fileHashByte = append(fileHashByte, headContentNodeUUIDByte...)
	fileHashByte = append(fileHashByte, tailContentNodeUUIDByte...)

	iv := userlib.RandomBytes(16)

	file.hashCheck = userlib.Hash(fileHashByte)
	file.EncryptedHashCheck = userlib.SymEnc(fileAccessNode.symmetricEncryptionKey, iv, file.hashCheck)
	// store file
	fileByte, err := json.Marshal(file)
	if err != nil {
		return err
	}
	userlib.DatastoreSet(fileAccessNode.fileUUID, fileByte)
	return nil
}

// get contentnode according to its uuid and a fileaccessnode
func (userdata *User) getContentNode(contentNodeUUID uuid.UUID, fileAccessNode FileAccessNode) (contentNode ContentNode, err error) {
	//try to get content node
	contentNodeByte, ok := userlib.DatastoreGet(contentNodeUUID)

	if !ok {
		return contentNode, errors.New("getContentNode an error occurred 1")
	}
	err = json.Unmarshal(contentNodeByte, &contentNode)

	if err != nil {
		return contentNode, errors.New("getContentNode an error occurred 2")
	}

	//check that contentNode isn't modified
	selfUUIDByte, err := json.Marshal(contentNode.SelfUUID)

	if err != nil {
		return contentNode, errors.New("getContentNode an error occurred 3")
	}

	nextUUIDByte, err := json.Marshal(contentNode.NextUUID)

	if err != nil {
		return contentNode, errors.New("getContentNode an error occurred 4")
	}

	var contentNodeHashByte = contentNode.EncryptedContent
	contentNodeHashByte = append(contentNodeHashByte, selfUUIDByte...)
	contentNodeHashByte = append(contentNodeHashByte, nextUUIDByte...)

	localHashCheck := userlib.Hash(contentNodeHashByte)

	contentNode.hashCheck = userlib.SymDec(fileAccessNode.symmetricEncryptionKey, contentNode.EncryptedHashCheck)

	ok = userlib.HMACEqual(localHashCheck, contentNode.hashCheck)

	if !ok {
		return contentNode, errors.New("getContentNode an error occurred 5")
	}

	//decrypt content
	contentNode.content = userlib.SymDec(fileAccessNode.symmetricEncryptionKey, contentNode.EncryptedContent)

	return contentNode, nil
}

// create new content Node
func (userdata *User) newContentNode(fileAccessNode FileAccessNode, content []byte, newUUID uuid.UUID) (err error) {
	var newContentNode ContentNode
	// set uuid
	newContentNodeUUID := newUUID
	// reset file tail uuid

	newContentNode.SelfUUID = newContentNodeUUID
	newContentNode.NextUUID = uuid.Nil
	newContentNode.content = content
	iv := userlib.RandomBytes(16)
	newContentNode.EncryptedContent = userlib.SymEnc(fileAccessNode.symmetricEncryptionKey, iv, content)

	//initialize content node's hash check
	selfUUIDByte, err := json.Marshal(newContentNode.SelfUUID)
	if err != nil {
		return errors.New("newContentNode an error occurred 2")
	}
	nextUUIDByte, err := json.Marshal(newContentNode.NextUUID)
	if err != nil {
		return errors.New("newContentNode an error occurred 3")
	}
	var newContentNodeHashByte = newContentNode.EncryptedContent
	newContentNodeHashByte = append(newContentNodeHashByte, selfUUIDByte...)
	newContentNodeHashByte = append(newContentNodeHashByte, nextUUIDByte...)

	newContentNode.hashCheck = userlib.Hash(newContentNodeHashByte)
	iv = userlib.RandomBytes(16)
	newContentNode.EncryptedHashCheck = userlib.SymEnc(fileAccessNode.symmetricEncryptionKey, iv, newContentNode.hashCheck)

	// store the content node dataStore
	newContentNodeByte, err := json.Marshal(newContentNode)
	if err != nil {
		return errors.New("newContentNode an error occurred 4")
	}
	userlib.DatastoreSet(newContentNode.SelfUUID, newContentNodeByte)

	return nil
}
func (userdata *User) updateContentNode(fileAccessNode FileAccessNode, nextContentNodeUUID uuid.UUID) (err error) {
	var file File

	file, err = userdata.getFile(fileAccessNode)

	if err != nil {
		return errors.New("updateContentNode an error occurred 1")
	}

	// reset pre content node uuid if
	var preContentNode ContentNode

	preContentNode, err = userdata.getContentNode(file.TailContentNodeUUID, fileAccessNode)
	if err != nil {
		return errors.New("updateContentNode an error occurred 2")
	}

	// reset pre content node

	preContentNode.NextUUID = nextContentNodeUUID
	iv := userlib.RandomBytes(16)
	preContentNode.EncryptedContent = userlib.SymEnc(fileAccessNode.symmetricEncryptionKey, iv, preContentNode.content)
	preContentNodeUUIDByte, err := json.Marshal(preContentNode.SelfUUID)
	if err != nil {
		return errors.New("updateContentNode an error occurred 3")
	}
	preContentNodenextUUIDByte, err := json.Marshal(preContentNode.NextUUID)
	if err != nil {
		return errors.New("updateContentNode an error occurred 4")
	}

	var preContentNodeHashByte = preContentNode.EncryptedContent
	preContentNodeHashByte = append(preContentNodeHashByte, preContentNodeUUIDByte...)
	preContentNodeHashByte = append(preContentNodeHashByte, preContentNodenextUUIDByte...)

	preContentNode.hashCheck = userlib.Hash(preContentNodeHashByte)
	iv = userlib.RandomBytes(16)
	preContentNode.EncryptedHashCheck = userlib.SymEnc(fileAccessNode.symmetricEncryptionKey, iv, preContentNode.hashCheck)

	// store the content node dataStore
	preContentNodeByte, err := json.Marshal(preContentNode)
	if err != nil {
		return errors.New("updateContentNode an error occurred 5")
	}
	// userlib.DatastoreDelete(preContentNode.SelfUUID)
	userlib.DatastoreSet(preContentNode.SelfUUID, preContentNodeByte)
	return nil
}

func (userdata *User) StoreFile(filename string, content []byte) (err error) {
	userdata.fileNameSpace, err = userdata.getFileNameSpace()
	if err != nil {
		return errors.New("StoreFile an error occurred 1")
	}

	fileAccessNodeUUID, ok := userdata.fileNameSpace.fileNameSpace[filename]
	var fileAccessNode FileAccessNode
	var file File

	// filename exist in file name space, get the file access node and access file struct.
	// else create new file access node and map it to file name space, create new file struct also
	if ok {
		fileAccessNode, err = userdata.getFileAccessNode(fileAccessNodeUUID)
		if err != nil {
			return errors.New("StoreFile an error occurred 2")
		}
		file, err = userdata.getFile(fileAccessNode)
		if err != nil {
			return errors.New("StoreFile an error occurred 3")
		}
	} else {
		/*initialize new file access node since no file name exist*/

		//initialize new file access node
		fileAccessNode, err = userdata.newFileAccessNode(userdata.Username, userlib.RandomBytes(16), uuid.New())
		if err != nil {
			return errors.New("StoreFile an error occurred 4")
		}
		fileAccessNodeUUID = fileAccessNode.SelfUUID

		userdata.fileNameSpace.fileNameSpace[filename] = fileAccessNodeUUID

		/*Initialize new file struct*/
		file.ArthorizedUsers = make(map[string]string)
		file.ArthorizedAccess = make(map[string]uuid.UUID)
		file.ArthorizedUsers[userdata.Username] = userdata.Username
		file.ArthorizedAccess[userdata.Username] = fileAccessNode.SelfUUID
	}

	// add new node
	newNodeUUID := uuid.New()

	file.HeadContentNodeUUID = newNodeUUID
	file.TailContentNodeUUID = newNodeUUID

	err = userdata.updateFile(fileAccessNode, file)
	if err != nil {
		return errors.New("StoreFile an error occurred 5")
	}

	err = userdata.newContentNode(fileAccessNode, content, newNodeUUID)
	if err != nil {
		return errors.New("StoreFile an error occurred 6")
	}

	//userdataBytes, err = json.Marshal(userdata)

	//userlib.DatastoreSet(userUUID, userdataBytes)
	userdata.updateUserFileNameSpace()
	return nil
}

func (userdata *User) updateUserFileNameSpace() (err error) {
	/*FileNameSpaceByte, err := json.Marshal(userdata.fileNameSpace)
	if err != nil {
		return errors.New("an error occurred")
	}
	userdata.EncryptedFileNameSpace = userlib.SymEnc(userdata.fly_key_1[32:48], userdata.Salt[0:16], FileNameSpaceByte)

	//Hash the above content and make sure the struct isn't modified by malicious action
	//HashByte := append([]byte(userdata.Username), userdata.Password, userdata.Salt, userdata.EncryptedPKEDecKey, userdata.EncryptedDSSIgnKey, userdata.EncryptedFileNameSpace)
	HashByte := []byte(userdata.Username)
	HashByte = append(HashByte, userdata.Password...)
	HashByte = append(HashByte, userdata.Salt...)
	HashByte = append(HashByte, userdata.EncryptedPKEDecKey...)
	HashByte = append(HashByte, userdata.EncryptedDSSignKey...)
	HashByte = append(HashByte, userdata.EncryptedFileNameSpace...)
	HashByte = append(HashByte, userdata.fly_key_1[48:63]...)
	userdata.HashCheck, err = userlib.HashKDF(userlib.Hash(HashByte)[:16], []byte("Check Not Modified"))

	userdataBytes, err := json.Marshal(userdata)
	userlib.DatastoreSet(userUUID, userdataBytes)*/

	fileNameSpaceByte, err := json.Marshal(userdata.fileNameSpace.fileNameSpace)
	if err != nil {
		return errors.New("an error occurred")
	}

	userdata.fileNameSpace.EncryptedFileNameSpace = userlib.SymEnc(userdata.fly_key_1[32:48], userdata.Salt[0:16], fileNameSpaceByte)

	HashByte := append(userdata.fileNameSpace.EncryptedFileNameSpace, userdata.fly_key_2[0:16]...)
	userdata.fileNameSpace.HashCheck, err = userlib.HashKDF(userlib.Hash(HashByte)[:16], []byte("Check Not Modified"))
	if err != nil {
		return errors.New("an error occurred")
	}

	newFileNameSpaceByte, err := json.Marshal(userdata.fileNameSpace)
	if err != nil {
		return errors.New("an error occurred")
	}

	userlib.DatastoreSet(userdata.FileNameSpaceUUID, newFileNameSpaceByte)

	return nil
}

func (userdata *User) AppendToFile(filename string, content []byte) error {
	//try retreive file access node in user's filenamespace
	FileNameSpace, err := userdata.getFileNameSpace()
	userdata.fileNameSpace = FileNameSpace
	if err != nil {
		return errors.New("AppendToFile an error occurred 1")
	}
	fileAccessNodeUUID, ok := userdata.fileNameSpace.fileNameSpace[filename]

	if !ok {
		return errors.New("AppendToFile an error occurred 2")
	}

	//get file access node
	fileAccessNode, err := userdata.getFileAccessNode(fileAccessNodeUUID)

	if err != nil {
		//return errors.New("AppendToFile an error occurred 3")
		return err
	}
	//get file

	file, err := userdata.getFile(fileAccessNode)

	if err != nil {
		//return errors.New("AppendToFile an error occurred 4")
		return err
	}

	newNodeUUID := uuid.New()
	err = userdata.updateContentNode(fileAccessNode, newNodeUUID)

	if err != nil {
		return errors.New("AppendToFile an error occurred 5")
	}

	err = userdata.newContentNode(fileAccessNode, content, newNodeUUID)
	if err != nil {
		return errors.New("AppendToFile an error occurred 6")
	}

	file.TailContentNodeUUID = newNodeUUID
	userdata.updateFile(fileAccessNode, file)
	return nil
}
func (userdata *User) LoadFile(filename string) (content []byte, err error) {
	/*FileAccessNodeUUID, ok := userdata.fileNameSpace[filename]
	// not exit return
	if !ok {
		panic(err)
	}
	//get file accessnode*/
	//try retreive file access node in user's filenamespace
	userdata.fileNameSpace, err = userdata.getFileNameSpace()
	if err != nil {
		return nil, errors.New("LoadFile an error occurred 1")
	}
	fileAccessNodeUUID, ok := userdata.fileNameSpace.fileNameSpace[filename]

	if !ok {
		return nil, errors.New("LoadFile an error occurred 1")
	}

	//get file access node
	fileAccessNode, err := userdata.getFileAccessNode(fileAccessNodeUUID)

	if err != nil {
		return nil, errors.New("LoadFile an error occurred 2")
	}
	//get file
	file, err := userdata.getFile(fileAccessNode)

	if err != nil {
		return nil, err //errors.New("LoadFile an error occurred 3")
	}
	var currentNode ContentNode
	for currentNodeUUID := file.HeadContentNodeUUID; currentNodeUUID != uuid.Nil; currentNodeUUID = currentNode.NextUUID {
		currentNode, err = userdata.getContentNode(currentNodeUUID, fileAccessNode)
		if err != nil {
			return nil, errors.New("LoadFile an error occurred 4")
		}
		content = append(content, currentNode.content...)
	}

	return content, nil
}

func (userdata *User) CreateInvitation(filename string, recipientUsername string) (
	invitationPtr uuid.UUID, err error) {
	// check recipient User exist
	userUUID, err := uuid.FromBytes(userlib.Hash([]byte(recipientUsername))[:16])
	if err != nil {
		return uuid.Nil, errors.New("CreateInvitation an error occurred 1")
	}

	user, ok := userlib.DatastoreGet(userUUID)
	_ = user
	if !ok {
		return uuid.Nil, errors.New("CreateInvitation an error occurred 2")
	}

	// check file name exist
	FileNameSpace, err := userdata.getFileNameSpace()
	fileName := FileNameSpace.fileNameSpace[filename]
	if err != nil {
		return uuid.Nil, errors.New("CreateInvitation an error occurred 3")
	}
	fileAccessNode, err := userdata.getFileAccessNode(fileName)
	if err != nil {
		return uuid.Nil, errors.New("CreateInvitation an error occurred 4")
	}
	// create new FileAcessNode

	newfileAccessNode, err := userdata.newFileAccessNode(recipientUsername, fileAccessNode.symmetricEncryptionKey, fileAccessNode.fileUUID)
	if err != nil {
		return uuid.Nil, errors.New("CreateInvitation an error occurred 5")
	}

	file, err := userdata.getFile(fileAccessNode)

	file.ArthorizedUsers[recipientUsername] = userdata.Username
	file.ArthorizedAccess[recipientUsername] = newfileAccessNode.SelfUUID
	err = userdata.updateFile(fileAccessNode, file)
	return newfileAccessNode.SelfUUID, nil
}

func (userdata *User) AcceptInvitation(senderUsername string, invitationPtr uuid.UUID, filename string) error {

	FileNameSpace, err := userdata.getFileNameSpace()
	var fileAccessNode FileAccessNode
	if err != nil {
		return errors.New("accept 1")
	}

	
	//get file access node and initialize its hash, then save it back to datastore
	fileAccessNodeByte, ok := userlib.DatastoreGet(invitationPtr)

	if !ok {
		return errors.New("accept 3")
	}
	err = json.Unmarshal(fileAccessNodeByte, &fileAccessNode)

	if err != nil {
		return errors.New("accept 4")
	}

	//compute fileUUID
	fileUUIDByte, err := userlib.PKEDec(userdata.pKEDecKey, fileAccessNode.EncryptedFileUUID)
	if err != nil {
		return errors.New("getFileAccessNode an error occurred 6")
	}

	err = json.Unmarshal(fileUUIDByte, &fileAccessNode.fileUUID)
	if err != nil {
		return errors.New("getFileAccessNode an error occurred 7")
	}

	// The caller already has a file with the given filename in their personal file namespace.
	fileAccessNodeUUID, ok := userdata.fileNameSpace.fileNameSpace[filename]
	if ok {
		fileAccessNodeB, err := userdata.getFileAccessNode(fileAccessNodeUUID)
		if err == nil {
			_, err:= userdata.getFile(fileAccessNodeB)
			if err == nil {
				if fileAccessNodeB.fileUUID != fileAccessNode.fileUUID {
					return errors.New("accept 2")
				}
			}
		}
	}
	
	userPKEEncryptKey, ok := userlib.KeystoreGet(userdata.Username + "PKEEncKey")
	if !ok {
		return errors.New("accept 5")
	}

	selfUUIDByte, err := json.Marshal(fileAccessNode.SelfUUID)
	if err != nil {
		return errors.New("accept 6")
	}

	FileAccessNodeHashByte := selfUUIDByte
	FileAccessNodeHashByte = append(FileAccessNodeHashByte, fileAccessNode.EncryptedFileUUID...)
	FileAccessNodeHashByte = append(FileAccessNodeHashByte, fileAccessNode.EncryptedSymmetricEncryptionKey...)
	FileAccessNodeHashByte = append(FileAccessNodeHashByte, userdata.fly_key_2[16:32]...)
	fileAccessNode.hashCheck = userlib.Hash(FileAccessNodeHashByte)
	fileAccessNode.EncryptedHashCheck, err = userlib.PKEEnc(userPKEEncryptKey, fileAccessNode.hashCheck)
	if err != nil {
		return errors.New("CreateInvitation an error occurred 1")
	}

	fileAccessNodeByte, err = json.Marshal(fileAccessNode)
	if err != nil {
		return errors.New("CreateInvitation an error occurred 1")
	}
	userlib.DatastoreSet(fileAccessNode.SelfUUID, fileAccessNodeByte)


	FileNameSpace.fileNameSpace[filename] = invitationPtr

	userdata.fileNameSpace = FileNameSpace
	userdata.updateUserFileNameSpace()

	return nil
}

func (userdata *User) RevokeAccess(filename string, recipientUsername string) error {
	var err error
	userdata.fileNameSpace, err = userdata.getFileNameSpace()
	if err != nil {
		return errors.New("RevokeAccess an error occurred 2")
	}

	fileAccessNodeUUID, ok := userdata.fileNameSpace.fileNameSpace[filename]

	if !ok {
		return errors.New("RevokeAccess an error occurred 3")
	}

	//get file access node
	fileAccessNode, err := userdata.getFileAccessNode(fileAccessNodeUUID)
	if err != nil {
		return errors.New("RevokeAccess an error occurred 4")
	}

	file, err := userdata.getFile(fileAccessNode)
	if err != nil {
		return errors.New("RevokeAccess an error occurred 5")
	}

	//download contents
	content, err := userdata.LoadFile(filename)
	if err != nil {
		return errors.New("RevokeAccess an error occurred 6")
	}



	//do the actual revoking
	_, ok = file.ArthorizedUsers[recipientUsername]
	if !ok {
		return errors.New("no such recipent")
	}

	var UsersToBeRemoved []string
	var Queue []string

	Queue = append(Queue, recipientUsername)

	for ;(len(Queue) != 0); {
		UserToBeRemoved := Queue[0]
		UsersToBeRemoved = append(UsersToBeRemoved, UserToBeRemoved)
		for accepter, inviter := range file.ArthorizedUsers {
			if inviter == UserToBeRemoved {
				Queue = append(Queue, accepter)
			}
		}
		Queue = Queue[1:]
	} 

	for i := range UsersToBeRemoved {
		delete(file.ArthorizedUsers, UsersToBeRemoved[i])
		delete(file.ArthorizedAccess, UsersToBeRemoved[i])
	}

	//compute new key to access file
	newSymmetricEncryptionKey := userlib.RandomBytes(16)

	//share new key
	for username, userFileAccessNodeUUID := range file.ArthorizedAccess {
		userPKEEncryptKey, ok:= userlib.KeystoreGet(username + "PKEEncKey")
		if !ok {
			return errors.New(username)
		}
		NewEncryptedSymmetricEncryptionKey, err := userlib.PKEEnc(userPKEEncryptKey, newSymmetricEncryptionKey)
		if err != nil {
			return err
		}
		DSSignNewEncryptedSymmetricEncryptionKey, err := userlib.DSSign(userdata.dSSignKey, NewEncryptedSymmetricEncryptionKey)
		if err != nil {
			return err
		}

		//get their file access node
		var userFileAccessNode FileAccessNode
		userFileAccessNodeByte, ok := userlib.DatastoreGet(userFileAccessNodeUUID)
		if !ok {
			return errors.New("getFileAccessNode an error occurred 8")
		}
		err = json.Unmarshal(userFileAccessNodeByte, &userFileAccessNode)
		if err != nil {
			return err
		}

		userFileAccessNode.NewSymmetricEncryptionKeyProvider = userdata.Username
		userFileAccessNode.NewEncryptedSymmetricEncryptionKey = NewEncryptedSymmetricEncryptionKey
		userFileAccessNode.DSSignNewSymmetricEncryptionKey = DSSignNewEncryptedSymmetricEncryptionKey

		//save the file access node
		fileAccessNodeByte, err := json.Marshal(fileAccessNode)
		if err != nil {
			return  err
		}

		userlib.DatastoreSet(fileAccessNode.SelfUUID, fileAccessNodeByte)
	}

	//refresh fileAccessNode
	fileAccessNode, err = userdata.getFileAccessNode(fileAccessNodeUUID)
	if err != nil {
		return errors.New("RevokeAccess an error occurred 9")
	}	

	//create new content node using the new file access node
	newContentNodeUUID := uuid.New()
	err = userdata.newContentNode(fileAccessNode, content, newContentNodeUUID)
	if err != nil {
		return errors.New("RevokeAccess an error occurred 10")
	}	

	//update file
	file.HeadContentNodeUUID = newContentNodeUUID
	file.TailContentNodeUUID = newContentNodeUUID

	err = userdata.updateFile(fileAccessNode, file)
	if err != nil {
		return errors.New("RevokeAccess an error occurred 11")
	}

	return nil
}
