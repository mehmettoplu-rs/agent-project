package tools

import "os"

// --- Read File Tool ---
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return `TOOL: read_file (To read a file from the system)
    {
        "thinking": "I need to read the contents of this specific file.",
        "action": "read_file",
        "data": "example.txt",
        "content": ""
    }`
}

func (t *ReadFileTool) Execute(data, content string) string {
	return "SYSTEM DATA (File Content of " + data + "):\n" + readFile(data) + "\nDO NOT call read_file again for this file. Reply to the user."
}

// --- Write File Tool ---
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return `TOOL: write_file (To create or overwrite a file)
    {
        "thinking": "I need to write this exact content to the file.",
        "action": "write_file",
        "data": "main.py",
        "content": "print('Hello World')"
    }`
}

func (t *WriteFileTool) Execute(data, content string) string {
	return "SYSTEM DATA: " + writeFile(data, content) + " Inform the user that the file has been created."
}

// --- Helper Functions ---
func readFile(filePath string) string {
	// os.ReadFile reads the entire file into memory at once.
	// For very large files, this allocates a massive byte slice on the heap, risking high GC pressure.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "ERROR: File could not be read - " + err.Error()
	}
	// string(data) converts the byte slice to a string. This operation creates a complete,
	// fresh copy of the data in memory because strings in Go are immutable.
	return string(data)
}

func writeFile(filePath, content string) string {
	// []byte(content) converts the string to a byte slice. Like the above, this creates
	// a full copy of the data in memory. os.WriteFile then writes it to disk in one shot.
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "ERROR: Could not write to the file - " + err.Error()
	}
	return "SUCCESS: File was written successfully."
}
