# Selpg.go

<!-- TOC -->

- [Selpg.go](#selpggo)
  - [设计说明](#设计说明)
    - [包引用](#包引用)
    - [初始化](#初始化)
    - [标识、参数预处理](#标识参数预处理)
    - [接受输入数据](#接受输入数据)
    - [数据处理](#数据处理)
    - [输出结果](#输出结果)
  - [使用方法](#使用方法)
    - [安装程序](#安装程序)
    - [使用范例和参数说明](#使用范例和参数说明)
      - [必需参数：-sNumber，-eNumber](#必需参数-snumber-enumber)
      - [互斥的可选参数：-lNumber，-f](#互斥的可选参数-lnumber-f)
      - [可选参数：-dDestination](#可选参数-ddestination)
      - [可选参数：file_name](#可选参数file_name)
  - [测试](#测试)
    - [生成测试文件](#生成测试文件)
    - [测试与结果](#测试与结果)

<!-- /TOC -->

## 设计说明

本程序参照 [开发 Linux 命令行实用程序](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html) 的设计，以 go 语言替代 C 语言构建。下面我们分版块来讲解实现。[完整源代码](https://github.com/SiskonEmilia/Selpg/blob/master/selpg.go)

### 包引用

```go
package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	flag "github.com/spf13/pflag"
)
```

在这一部分，我们声明了程序所要用到的所有包：

- `bufio`：用于从标准输入流获取数据和将数据写入到标准输出流
- `io`：用于引用 `io.EOF` 来判断错误是否是文件尾导致
- `log`：用于将错误信息写入到标准错误流
- `os`：用于打开文件和异常退出时发送状态码
- `os/exec`：用于开启 `lp` 子进程
- `strings`：用于划分、拼接字符串
- `github.com/spf13/pflag`：用于获取程序运行时用户输入的参数和标识

### 初始化

```go
// Initializing //
startNumber := flag.IntP("startpage", "s", 0, "The page to start printing at [Necessary, no greater than endpage]")
endNumber := flag.IntP("endpage", "e", 0, "The page to end printing at [Necessary, no less than startpage]")
lineNumber := flag.IntP("linenumber", "l", 72, "If this flag is used, a page will consist of a fixed number of characters, which is given by you")
forcePage := flag.BoolP("forcepaging", "f", false, "Change page only if '-f' appears [Cannot be used with -l]")
destinationPrinter := flag.StringP("destination", "d", "", "Choose a printer to accept the result as a task")

// StdErr printer //
l := log.New(os.Stderr, "", 0)

// Data holder //
bytes := make([]byte, 65535)
var data string
var resultData string

flag.Parse()
```

在这个部分，我们进行了初始化操作。这包括对于 `pflag` 中各个标识的设置和变量绑定，标准错误流的绑定，缓冲区 `bytes` 的创建，读入数据变量、结果数据变量的创建。

在完成这些设置后，我们通过 `flag.Parse()` 方法使得 `pflag` 执行对于标识和参数的解析。

### 标识、参数预处理

```go
// Are necessary flags given? //
if *startNumber == 0 || *endNumber == 0 {
  l.Println("Necessary flags are not given!")
  flag.Usage()
  os.Exit(1)
}

// Are flags value valid? //
if (*startNumber > *endNumber) || *startNumber < 0 || *endNumber < 0 || *lineNumber <= 0 {
  l.Println("Invalid flag values!")
  flag.Usage()
  os.Exit(1)
}

// Are lineNumber and forcePage set at the same time? //
if *lineNumber != 72 && *forcePage {
  l.Println("Linenumber and forcepaging cannot be set at the same time!")
  flag.Usage()
  os.Exit(1)
}

// Too many arguments? //
if flag.NArg() > 1 {
  l.Println("Too many arguments!")
  flag.Usage()
  os.Exit(1)
}
```

在这部分，我们检验了所有标识的合法性，这包括：

- 必须的标识，`-s` 和 `-e` 是否被设置？
- 标识是否具有一个合法的值
- 互斥的参数，也就是通过行数分页和通过分页符分页，是否被同时设置
- 参数数量是否过多

如果任何不合法的参数被使用，那么我们向标准错误流输出错误信息，打印正确使用方法，然后退出程序（并返回一个通用的错误状态码）。

### 接受输入数据

```go
// StdIn or File? //
if flag.NArg() == 0 {
  // StdIn condition //
  reader := bufio.NewReader(os.Stdin)

  size, err := reader.Read(bytes)

  for size != 0 && err == nil {
    data = data + string(bytes)
    size, err = reader.Read(bytes)
  }

  // Error
  if err != io.EOF {
    l.Println("Error occured when reading from StdIn:\n", err.Error())
    os.Exit(1)
  }

} else {
  // File condition //
  file, err := os.Open(flag.Args()[0]) // TODO TEST: is PATH needed?
  if err != nil {
    l.Println("Error occured when opening file:\n", err.Error())
    os.Exit(1)
  }

  // Read the whole file
  size, err := file.Read(bytes)

  for size != 0 && err == nil {
    data = data + string(bytes)
    size, err = file.Read(bytes)
  }

  // Error
  if err != io.EOF {
    l.Println("Error occured when reading file:\n", err.Error())
    os.Exit(1)
  }
}
```

在这一部分，我们判断输入方式，并且将数据读入并写在 `data` 变量中。

对于 **标准输入** 的模式，也就是没有额外参数的情况：我们首先通过 `bufio.NewReader(os.Stdin)` 创建一个绑定到标准输入流的读者，然后通过它向缓冲区 `bytes` 读入数据，并且将其转换为字符串并写入到 `data` 中。由于缓冲区的大小限制，这个读入过程可能需要进行多次，因而我们迭代该过程，直到确保读完了标准输入流的数据（也就是该次读入没有读入到数据，即 `size = 0`）为止。在读入遇到错误时，我们输出错误信息，并且退出程序。

对于 **文件输入** 的模式，也就是有一个参数的情况：我们首先通过 `os.Open()` 打开文件。在没有错误的情况下，我们通过 `file.Read()` 迭代地从中读入数据，直到完成读取。如果我们在整个过程中遇到错误，那么输出错误信息，并且退出程序。

在完成这一个部分的处理后，我们的数据信息就存储在了 `data` 变量中。

### 数据处理

```go
// LineNumber or ForcePaging? //
if *forcePage {
  // ForcePaging //
  pagedData := strings.SplitAfter(data, "\f")

  if len(pagedData) < *endNumber {
    l.Println("Invalid flag values! Too large endNumber!")
    flag.Usage()
    os.Exit(1)
  }

  resultData = strings.Join(pagedData[*startNumber-1:*endNumber+1], "")
} else {
  // LineNumber //
  lines := strings.SplitAfter(data, "\n")

  if len(lines) < (*endNumber-1)*(*lineNumber)+1 {
    l.Println("Invalid flag values! Too large endNumber!")
    flag.Usage()
    os.Exit(1)
  }
  if len(lines) < *endNumber*(*lineNumber) {
    resultData = strings.Join(lines[(*startNumber)*(*lineNumber)-(*lineNumber):len(lines)], "")
  } else {
    resultData = strings.Join(lines[(*startNumber)*(*lineNumber)-(*lineNumber):(*endNumber)*(*lineNumber)], "")
  }
}
```

在这部分，我们对存储在 `data` 里的字符串进行处理，以满足用户要求。这部分根据分页过程的不同分为两种：按分页符分页和按行分页。

在按 **分页符分页** 的情况下，我们通过 `strings.SplitAfter()` 方法来将字符串以 `'\f'` 为界分为数个段，每一段即是一页，然后我们根据用户输入的开始页码和结束页码将相应的数据写入 `resultData` 中。

在 **按行数分页** 的情况下，我们首先以 `'\n'` 为界将字符串分段，然后根据开始页码和结束页码计算出开始行和结束行，并将其间数据写入 `resultData` 中。

### 输出结果

```go
writer := bufio.NewWriter(os.Stdout)

// StdOut or Printer? //
if *destinationPrinter == "" {
  // StdOut //
  fmt.Printf("%s", resultData)
} else {
  // Printer //
  cmd := exec.Command("lp", "-d"+*destinationPrinter)
  lpStdin, err := cmd.StdinPipe()

  if err != nil {
    l.Println("Error occured when trying to send data to lp:\n", err.Error())
    os.Exit(1)
  }
  go func() {
    defer lpStdin.Close()
    io.WriteString(lpStdin, resultData)
  }()

  out, err := cmd.CombinedOutput()
  if err != nil {
    l.Println("Error occured when sending data to lp:\n", err.Error())
    os.Exit(1)
  }

  _, err = writer.Write(out)

  if err != nil {
    l.Println("Error occured when writing information to StdOut:\n", err.Error())
    os.Exit(1)
  }
}
```

在这部分，我们以输出方式的不同分为两类：直接输出到标准输出流的和将数据传送给 `lp` 进行打印工作的。由于二者实际上都需要用到标准输出流（后者是要输出 `lp` 的信息），所以我们首先创建了与标准输出流绑定的 `Writer`：`writer`。

对于直接输出到标准输出流的，我们直接通过 `Writer.Write()` 方法将转换为 byte 切片的字符串输出即可。

而对于输出到 `lp` 的，我们首先通过 `exec.Command()` 创建一个 `lp` 的子进程，并且通过 `cmd.StdinPipe()` 获取和其标准输入绑定的管道，然后将数据送入管道即可。同时，我们也需要将 `lp` 指令的输出转发到标准输出流上，方便用户查看。

在这期间对于错误的处理依旧和前文相同：输出错误并退出程序。

## 使用方法

使用方法基本同 C 版本的 `selpg`。

### 安装程序

在配置好 golang 环境的前提下，运行：

```bash
go get github.com/siskonemilia/selpg
```

若成功执行（无回显），则安装成功。

### 使用范例和参数说明

```bash
selpg -sNumber -eNumber [-lNumber/-f] [-dDestination] [file_name]
```

#### 必需参数：-sNumber，-eNumber

`selpg` 要求用户用两个命令行参数“-sNumber”（例如，“-s10”表示从第 10 页开始）和“-eNumber”（例如，“-e20”表示在第 20 页结束）指定要抽取的页面范围的起始页和结束页。`selpg` 对所给的页号进行合理性检查；换句话说，它会检查两个数字是否为有效的正整数以及结束页是否不小于起始页。两者是程序执行所必需的。

#### 互斥的可选参数：-lNumber，-f

selpg 可以处理两种输入文本：

**类型 1**：该类文本的页行数固定。这是缺省类型，因此不必给出选项进行说明。也就是说，如果既没有给出“-lNumber”也没有给出“-f”选项，则 `selpg` 会理解为页有固定的长度（每页 72 行）。例如：

    selpg -s10 -e20 -l66

**类型 2**：该类型文本的页由 ASCII 换页字符（十进制数值为 12，在 C 中用“\f”表示）定界。该格式与“每页行数固定”格式相比的好处在于，当每页的行数有很大不同而且文件有很多页时，该格式可以节省磁盘空间。在含有文本的行后面，类型 2 的页只需要一个字符 ― 换页 ― 就可以表示该页的结束。打印机会识别换页符并自动根据在新的页开始新行所需的行数移动打印头。例如：

    selpg -s10 -e20 -f 

#### 可选参数：-dDestination

`selpg` 还允许用户使用“-dDestination”选项将选定的页直接发送至打印机。这里，“Destination”应该是 lp 命令“-d”选项（请参阅“man lp”）可接受的打印目的地名称。该目的地应该存在 ― `selpg` 不检查这一点。在运行了带“-d”选项的 `selpg` 命令后，若要验证该选项是否已生效，请运行命令“lpstat -t”。该命令应该显示添加到“Destination”打印队列的一项打印作业。如果当前有打印机连接至该目的地并且是启用的，则打印机应打印该输出。

#### 可选参数：file_name

如果没有给出 file_name，那么 `selpg` 将从标准输入流读取数据进行处理，否则，`selpg` 将根据文件名寻找相应文件，并从中读取数据进行处理。

## 测试

该部分参照 [使用 selpg](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html#6) 进行程序测试。

### 生成测试文件

为了使得实验结果直观易懂，我们采用固定的程序生成的输入文件进行测试，生成程序如下：

```sh
# input_file_generator.sh
for i in {1..7200}
do
  echo $i >> input_file
done
```

其生成的结果是一个由 1 到 7200 的，步长为 1 的等差数列，且每行一个数字。即是说，每一行的数字都是该行的编号。

### 测试与结果

1. 把 `input_file` 的第 1 页写至标准输出

        $selpg -s1 -e1 input_file

    结果：

        1
        2
        ...
        72

2. `selpg` 读取标准输入，而标准输入已被 shell 重定向为来自 `input_file` 而不是显式命名的文件名参数。输入的第 1 页被写至屏幕

        $ selpg -s1 -e1 < input_file
    
    结果：

        1
        2
        ...
        72

3. `cat` 的标准输出被 shell／内核重定向至 selpg 的标准输入。将第 10 页到第 20 页写至 selpg 的标准输出

        $ cat input_file | selpg -s10 -e20
    
    结果：

        649
        650
        ...
        1440

4. `selpg` 将第 10 页到第 20 页写至标准输出（屏幕）；所有的错误消息被 shell 重定向至 `error_file`

        $ selpg -s10 -e20 input_file 2>error_file

    结果：

        649
        650
        ...
        1440
    
    `error_file`: 无内容

5. `selpg` 将第 10 页到第 20 页写至标准输出；标准输出被 shell 重定向至 `res`

        $selpg -s10 -e20 input_file >res

    结果（res文件内容）：

        649
        650
        ...
        1440

1. `selpg` 将第 10 页到第 20 页写至标准输出，标准输出被重定向至 `res`，`selpg` 写至标准错误的所有内容都被重定向至 `error_file`

        $selpg -s10 -e20 input_file >res 2>error_file

    结果（res文件内容）：

        649
        650
        ...
        1440
    
    `error_file`: 无内容

1. `selpg` 的标准输出透明地被 shell 重定向，成为 `cat` 的标准输入，第 10 页到第 20 页被写至该标准输入

        $ selpg -s10 -e20 input_file | cat

    结果：

        649
        650
        ...
        1440

1. 将页长设置为 66 行，这样 `selpg` 就可以把输入当作被定界为该长度的页那样处理。第 10 页到第 20 页被写至 `selpg` 的标准输出

        $ selpg -s10 -e20 -l66 input_file
    
    结果：

        595
        596
        ...
        1320

1. 假定页由换页符定界。第 10 页到第 20 页被写至 `selpg` 的标准输出

        $ selpg -s10 -e20 -f input_file

    结果（输入文件没有换页符，所以只有一页）：

        Invalid flag values! Too large endNumber!
        Usage of selpg:
        -d, --destination string   Choose a printer to accept the result as a task
        -e, --endpage int          The page to end printing at [Necessary, no less than startpage]
        -f, --forcepaging          Change page only if '-f' appears [Cannot be used with -l]
        -l, --linenumber int       If this flag is used, a page will consist of a fixed number of characters, which is given by you (default 72)
        -s, --startpage int        The page to start printing at [Necessary, no greater than endpage]

1. 第 10 页到第 20 页由管道输送至命令 `lp -dlp1`，该命令将使输出在打印机 lp1 上打印

        $ selpg -s10 -e20 -dlp1 input_file
    
    结果（因为没有打印机）：

        Error occured when sending data to lp:
        exit status 1

经测试，程序运行符合预期，工作正常。