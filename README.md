# 2024-06-07,20点25
研究go的源码 看的最新的1.22.4
我会把整个学习过程都写在这个readme.md里面.希望后续可以整理成博客或者书.

##首先是基础知识部分.
#plan 9汇编:


https://plan9.io/sys/doc/comp.pdf

How to Use the Plan 9 C Compiler
                        Rob Pike





# plan9基本知识:
	首先我们学习如何用vscode+delve来调试go的plan9汇编代码.
	首先我们用vscode配置好go的运行环境.能正确打印helloworld代码.
	编写代码:
	//main.go
	package main

	func main() {

		var aaa = Sum(2, 4)
		print(aaa)
	}

	func Sum(x, y int) int

	//add.s
	TEXT ·Sum(SB), $0-8
    MOVQ x+0(FP), AX  // 将第一个参数 x 放入 AX
    MOVQ y+8(FP), BX  // 将第二个参数 y 放入 BX
    ADDQ BX, AX       // 将 BX 加到 AX
    MOVQ AX, ret+16(FP)  // 将结果从 AX 移到返回值位置
    RET               // 返回
	之后我们sum这行在go代码里面打断点.单步调试就会发现delve调试器自动进入了汇编代码中.
	这时候我们在watch里面可以输入寄存器名字这些来查看寄存器.
	一些关键寄存器: RSP, RBP, RAX, RBX, RCX, RDX, RSI, RDI





	https://cloud.tencent.com/developer/article/2416368

	go汇编中出现unexpected EOF asm: assembly of pkg\test.s failed的解决办法
	这个bug的解决办法就是在go汇编代码最后换一行就行了


	https://blog.csdn.net/qq_17818281/article/details/114891093



	一套另外的中文教程: 非常好.
	https://golang.design/under-the-hood/zh-cn/part1basic/ch01basic/asm/




	一份很详细的博客:
	https://blog.csdn.net/zhu0902150102/article/details/129539307

  




Go语言汇编:
https://9p.io/sys/doc/asm.html      原始资料的地址.
Plan 9汇编
寄存器：
数据寄存器：R0-R7，地址寄存器：A0-A7，浮点寄存器：F0-F7。
A6指向数据.

伪栈寄存器：FP, SP, TOS。

FP是frame pointer，0(FP)是第一个参数，4(FP)是第二个。

SP是local stack pointer，保存自动变量。0(SP)是第一个。

TOS是top of stack寄存器，用来保存过程的参数，保存局部变量。

汇编器可以有一个变量名，比如p+0(FP)，表示p是第一个参数，这个变量保存在符号表内，但是对程序运行没有影响。实际有用的是 0(FP), 左边那个p只是给程序员看的. 对于计算机没用.但是要求必须写,

例子:下面2个代码都是等效的.都可以直接go run main.go
```
//add.s:
TEXT ·Sum(SB), $0-8
    MOVQ x+0(FP), AX  // 将第一个参数 x 放入 AX
    MOVQ y+8(FP), BX  // 将第二个参数 y 放入 BX
    ADDQ BX, AX       // 将 BX 加到 AX
    MOVQ AX, ret+16(FP)  // 将结果从 AX 移到返回值位置
    RET               // 返回
//main.go
package main

import "fmt"

func main() {
	x := 10
	y := 20
	sum := Sum(x, y)
	fmt.Println("Sum:", sum)
}

//go:noescape
func Sum(x, y int) int



```

```
//add.s:
TEXT ·Sum(SB), $0-8
    MOVQ x1213+0(FP), AX  // 将第一个参数 x 放入 AX //注意这里面xy的变量名,随便写.无所谓程序运行.
    MOVQ y2324+8(FP), BX  // 将第二个参数 y 放入 BX
    ADDQ BX, AX       // 将 BX 加到 AX
    MOVQ AX, ret+16(FP)  // 将结果从 AX 移到返回值位置
    RET               // 返回
//main.go
package main

import "fmt"

func main() {
	x := 10
	y := 20
	sum := Sum(x, y)
	fmt.Println("Sum:", sum)
}

//go:noescape
func Sum(x, y int) int

```


一些常用汇编指令表:
movb（8位）、movw（16位）、movl（32位）、movq（64位）//因为32位计算机, 指针是32位的所以我们读取变量使用movl

更复杂的go汇编指令可以参考这个网站,建议后续汇编代码遇到了回来查:
https://www.quasilyte.dev/blog/post/go-asm-complementary-reference/

比如movo= movdqa 移动双四字节 dq表示128位.所以movo表示对齐移动128位

AESENC: AESENC和AESENCLAST
这两条指令是AESNI中用于加密的指令，也是最容易理解的指令。任何SIMD指令都可以参考Intel® Intrinsics Guide，AESENC对输入依次进行ShiftRows，SubBytes，MixColumns，AddRoundKey操作。其中SubBytes是对字节的操作，因此可以和ShiftRows互换，与上面的图比较，可以发现AESENC恰好是上图的一个普通轮加密。

AESENCLAST对输入依次进行ShiftRows，SubBytes，AddRoundKey操作，相当于上图的尾轮加密。


32位系统和64位系统汇编函数传值的区别:
https://blog.csdn.net/song_lee/article/details/105297902
注意32位是用栈来传,64位是用寄存器来传值.所以同样一个代码64位比32位跑的快多了.

内存结构图:
<img src='huibian1.png'>
通过图很容易看出来ret的地址就是+16(fp)









数据：
所有的外部引用都需通过伪寄存器: PC（virtual Program Counter）/SB（Static Base register）。

PC用来控制程序执行
SB用来引用全局变量。

比如：

把全局数组的地址压栈：MOVL $array(SB), TOS。

把全局数组的第二个元素压栈：MOVL array+4(SB), TOS

local<>+4(SB)是本地变量，只在本文件可见。  <>表示局部变量. 不加<>表示外部变量.


bra: 把目的操作数传递到PC寄存器  bra: branch
    bra.w   $18000        *从10000跳转到18000

bsf bsr:
  格式: BSF dest, src
  影响标志位: ZF
  功能：从源操作数的的最低位向高位搜索，将遇到的第一个“1”所在的位序号存入目标寄存器中，
  若所有位都是0，则ZF=1，否则ZF=0。
  格式: BSR dest, src
  影响标志位: ZF
  功能：从源操作数的的最高位向低位搜索，将遇到的第一个“1”所在的位序号存入目标寄存器中，
  若所有位都是0，则ZF=1，否则ZF=0。

访问全局数据:
  MOVL $a6base(SB), A6      把SB放到A6寄存器里面.



放置数据:
  long word 放置他们的参数,放置适合的大小.
  比如:
    LONG $12345 把12345这个数据(base 10)放置到指令流里面.
  把数据放到data section里面复杂一点.
  指令DATA 接受2个参数, 第一个是放置数据的地址, 第二个是value
  例如:

      DATA  array+0(SB)/1,  $'a'   #$表示数据的意思.   /1表示放入的大小.
      DATA  array+1(SB)/1,  $'b'
      DATA  array+2(SB)/1,  $'c'
      GLOBL array(SB), $4          #4是ascii吗里面的结束符.

  或者 
      DATA  array+0(SB)/4,  $'abc\z'
      GLOBL array(SB), $4  
  解释: 
      GLOBL表示让这个符号array 变成全局变量. $4表示这个变量占用多少byte.


  DYNT INIT允许在Alef编译器上动态的类型.





FP: Frame pointer: arguments and locals.
PC: Program counter: jumps and branches.
SB: Static base pointer: global symbols.
SP: Stack pointer: the highest address within the local stack frame.

   the symbol foo(SB) is the name foo as an address in memory.
   foo(SB)表示SB里面开一个全局变量叫foo
   foo+4(SB) is four bytes past the start of foo.
   foo地址加4表示的地址. 并且这个地址在SB里面.
   Adding <> to the name, as in foo<>(SB), makes the name visible only in the current source file


   x-8(SP), y-4(SP):局部变量用sp, 因为栈是从大到小的. 所以是负号.因为之前x在sp中开变量,他肯定是往小方向走.

   每一个跳跃的label都只在他被定义的函数中才有效.
   所以多个函数中的label可以重名.


    In the general case, the frame size is followed by an argument size, separated by a minus sign. (It's not a subtraction, just idiosyncratic syntax.) The frame size $24-8 states that the function has a 24-byte frame and is called with 8 bytes of argument

    24-8表示栈大小24byte, 参数大小8byte



这是一个完整的函数定义demo:
TEXT runtime·profileloop(SB),NOSPLIT,$8
	MOVQ	$runtime·profileloop1(SB), CX
	MOVQ	CX, 0(SP)
	CALL	runtime·externalthreadhandler(SB)
	RET

 runtime·profileloop 函数名字 栈大小是8, 返回不写.
 把profileloop1函数放到cx里面, 再cx放到sp里面.这样栈里面就放入函数了.之后我们call就表示调用栈里面这个函数.最后ret即可. 这个代码就是调用其他函数.



全局变量赋值的语法:
DATA	symbol+offset(SB)/width, value



DATA divtab<>+0x00(SB)/4, $0xf4f8fcff  # 每4个位置进行一个赋值.
DATA divtab<>+0x04(SB)/4, $0xe6eaedf0
...
DATA divtab<>+0x3c(SB)/4, $0x81828384  
GLOBL divtab<>(SB), RODATA, $64   # 定义和初始化一个read only 变量. 这个变量是全局变量. 长度64

GLOBL runtime·tlsoffset(SB), NOPTR, $4



PCALIGN $32 # 下个命令对齐到32位.对齐好方便硬件加速.
MOVD $2, R0


go跟汇编的常量转化.
For example, given the Go declaration const bufSize = 1024, assembly code can refer to the value of this constant as const_bufSize.


type reader struct {
	buf [bufSize]byte
	r   int
}


Assembly can refer to the size of this struct as reader__size and the offsets of the two fields as reader_buf and reader_r. Hence, if register R1 contains a pointer to a reader, assembly can reference the r field as reader_r(R1).

汇编访问方法:
  reader__size   reader_buf  reader_r  reader_r(R1)


NOPTR 表示这个数据不会配置指针来表示. 他就是直接存数据的.


A good general rule of thumb is to define all non-RODATA symbols in Go instead of in assembly.

非只读数据用go代码来写.


如果函数没有参数和返回值. 写作$n-0.


 src/cmd/internal/obj/arm  这里面写了所有arm的指令集. 他们都以A为前缀.

386 amd在cmd/internal/obj/x86/a.out.go.

指令集都是从左到右的顺序:
  MOVQ $0, CX  表示clears CX



我们看64位的特殊地方:
  get_tls 是一个宏. 用来访问g和m指针.


  get_tls(CX)
  MOVQ	g(CX), AX     // Move g into AX.
  MOVQ	g_m(AX), BX   // Move g.m into BX.

  callee-save 可以长时间的数据
  caller-save 临时数据.



BYTE语法:byte就是往汇编里面放入机器码
Placing data in the instruction stream, say for interrupt vectors, is easy: the
pseudo-instructions LONG and WORD (but not BYTE) lay down the value of their
single argument, of the appropriate size, as if it were an instruction:
    LONG    $12345
places the long 12345 (base 10) in the instruction stream. (On most machines,
the only such operator is WORD and it lays down 32-bit quantities. The 386 has all
three: LONG, WORD, and BYTE.





MOV AX,BX 的二进制编码为:100010   011  000  0000
MOV AH, DL的二进制编码为:100010   110  110  1010
MOV AL, OFh 的二进制编码为:1011000001101111
MOV[BX],AX的二进制编码为:1000100010000000



```


下面我们再重头分析这个代码.
源码中的位置是src\runtime\internal\atomic\atomic_386.s
//这个后面是386.s所以是运行在32位系统的.
```
// uint64 atomicload64(uint64 volatile* addr);
TEXT ·Load64(SB), NOSPLIT, $0-12
	NO_LOCAL_POINTERS
	MOVL	ptr+0(FP), AX  //movb（8位）、movw（16位）、movl（32位）、movq（64位）//因为32位计算机, 指针是32位的所以我们读取变量使用movl
	TESTL	$7, AX
	JZ	2(PC)
	CALL	·panicUnaligned(SB)  //触发未对齐的错误.
	MOVQ	(AX), M0            //AX是表示他的值, (AX)是AX的值作为地址,来取这个地址的值.
	MOVQ	M0, ret+4(FP)  //这行命令需要结合main函数, main函数调用这个load64函数, 然后他先压入参数也就是uint64* addr. 一个8byte的指针, 然后压入load函数调用之后的下行位置.占用4字节. 所以我们代码第一行是0-12. 0表示(load64的)栈空间是0, 12是刚才的8+4. 之后我们算完M0,需要给ret+4的地址赋值. ret是返回位置也就是load函数调用之后的下行位置. 他加4,就是我们的返回值的指针位置即可.
	EMMS   //emms指令是在x86架构中用于清除MMX(多媒体扩展)寄存器状态的指令。
	RET

```




#复习汇编的函数调用:https://kakaroto.homelinux.net/2017/11/introduction-to-reverse-engineering-and-assembly/
int main() {
   return add_a_and_b(2, 3);
}

int add_a_and_b(int a, int b) {
   return a + b;
}


汇编:
_main:
   push   3                ; Push the second argument '3' into the stack
   push   2                ; Push the first argument '2' into the stack
   call   _add_a_and_b     ; Call the _add_a_and_b function. This will put the address of the next
                           ; instruction (add) into the stack, then it will jump into the _add_a_and_b
                           ; function by putting the address of the first instruction in the _add_a_and_b
                           ; label (push %ebx) into the EIP register
   add    %esp, 8          ; Add 8 to the esp, which effectively pops out the two values we just pushed into it
   ret                     ; Return to the parent function.... 

_add_a_and_b:
   push   %ebx             ; We're going to modify %ebx, so we need to push it to the stack
                           ; so we can restore its value when we're done
   mov    %eax, [%esp+8]   ; Move the first argument (8 bytes above the stack pointer) into EAX
   mov    %ebx, [%esp+12]  ; Move the second argument (12 bytes above the stack pointer) into EBX
   add    %eax, %ebx       ; Add EAX and EBX and store the result into EAX
   pop    %ebx             ; Pop EBX to restore its previous value
   ret                     ; Return back into the main. This will pop the value on the stack (which was
                           ; the address of the next instruction in the main function that was pushed into
                           ; the stack when the 'call' instruction was executed) into the EIP register


call xxx: 压入xxx的下一行位置.然后IP进入这个函数.
注意栈是从高到低.所以一个变量的起始坐标是她的下方!!!!!!!!!!!!!!!!!!!!!这点最重要对于栈理解!!!!!!!!!!!!!!!!!!!!!!!!!!
模拟整个过程:
  首先压入3,压入2.然后esp指向2的开始地址.继续续压入call的结束地址,继续压入ebx.
  这时2的开始地址(也就是2的地址小端,注意栈是从大到小生长的)比esp大8.因为中间夹着一个call的结束地址和一个ebx
  2放入eax, 3放入ebx, 然后结果加到eax.之后pop栈最后一个元素给ebx.最后esp加8.等于把栈清空了.
  返回值就是eax. 我理解pop    %ebx 好像没用.一会儿代码试试.也就是这样可以随便pop一下.
  




#我们继续读doc/go_mem.html
一个内存操作关注4个点:
  1.是读还是写,是原子操作,互斥操作,还是channel操作
  2.在程序中的位置
  3.他所访问的内存和变量
  4.操作读还是写的变量
如果p引入库包q, 那么q的初始化函数都在p的初始化之前.
所以所有函数初始化都在main之前.

```
var c = make(chan int, 10)
var a string

func f() {
	a = "hello, world"
	c <- 0
}

func main() {
	go f()
	<-c
	print(a)
}
```
这个代码可以保证打印hello,world, 把10丢掉也是正确的.无论带不带上10: 我们main函数需要c往外吐一个数才能启动, 但是c里面是空,所以f里面c<-0之后,print才启动.


```
var c = make(chan int)
var a string

func f() {
	a = "hello, world"
	<-c
}

func main() {
	go f()
	c <- 0
	print(a)
}
```
可以正确打印.因为c不带缓存,上来就是阻塞的.只有f里面<-c了,main才跑print



```
var c = make(chan int,10)
var a string

func f() {
	a = "hello, world"
	<-c
}

func main() {
	go f()
	c <- 0
	print(a)
}
```
不会打印hello,world.因为我们考虑c带缓存, 那么他就是上来就是非阻塞的.我们main走c<-0时候,f运行不运行都无所谓.所以大概率代码直接打印空就结束了.




var limit = make(chan int, 3)

func main() {
	for _, w := range work {
		go func(w func()) {
			limit <- 1
			w()
			<-limit
		}(w)
	}
	select{}
}
// work是一个任务组成的数组.
//select语句会一直监听所有指定的通道，直到其中一个通道准备好就会执行相应的代码块。
// 这里面一直是空, 所以整个程序阻塞.
//会把work里面的程序都起来, 但是limit大小是3.有3个在运行的w任务时候就会阻塞.这跟信号量效果一样.

锁:
sync里面有 sync.Mutex and sync.RWMutex.



var l sync.Mutex
var a string

func f() {
	a = "hello, world"
	l.Unlock()
}

func main() {
	l.Lock()
	go f()
	l.Lock()
	print(a)
}







Once:
var a string
var once sync.Once

func setup() {
	a = "hello, world"
}

func doprint() {
	once.Do(setup)
	print(a)
}

func twoprint() {
	go doprint()
	go doprint()
}


Atomic Values

Finalizers

Additional Mechanisms
  condition variables, lock-free maps, allocation pools, and wait groups. 

例子:
var a, b int

func f() {
	a = 1
	b = 2
}

func g() {
	print(b)
	print(a)
}

func main() {
	go f()
	g()
}
//it can happen that g prints 2 and then 0.
//main里面print b和a  同时f里面  a=1, b=2 有可能2赋值上了,1还没赋值上.但是我自己测试没复现出来.





var a string
var done bool

func setup() {
	a = "hello, world"
	done = true
}

func doprint() {
	if !done {
		once.Do(setup)
	}
	print(a)
}

func twoprint() {
	go doprint()
	go doprint()
} //没法保证能打印一次.







var a string
var done bool

func setup() {
	a = "hello, world"
	done = true
}

func main() {
	go setup()
	for !done {
	}
	print(a)
}//忙等也没法保证


type T struct {
	msg string
}

var g *T

func setup() {
	t := new(T)
	t.msg = "hello, world"
	g = t
}

func main() {
	go setup()
	for g == nil {
	}
	print(g.msg)
}// 也是错的





Not introducing data races into race-free programs means not moving writes out of conditional statements in which they appear. For example, a compiler must not invert the conditional in this program:
//程序1
*p = 1
if cond {
	*p = 2
}


//程序2
*p = 2
if !cond {
	*p = 1
}

If cond is false and another goroutine is reading *p, then in the original program, the other goroutine can only observe any prior value of *p and 1. In the rewritten program, the other goroutine can observe 2, which was previously impossible.

上面这2个程序不是等效的.
第一个程序我们*p=1, cond=false , 那么另外一个进程读了1
第二个程序我们cond=false, 但是另外一个进程读的快,把之间的2读走了.
所以这两个程序不等效.这就是并发的让问题变复杂了.原则就是不要在cond里面修改并发的变量.



n := 0
for e := list; e != nil; e = e.next {
	n++
}
i := *p
*q = 1
//list死循环, 那么也会在两个进程读写时候发生下面代码i := *p的运行.



*p = i + *p/2
//也不对



n := 0
for i := 0; i < m; i++ {
	n += *shared
}
into:
n := 0
local := *shared
for i := 0; i < m; i++ {
	n += local
}

这两种代码是等价的. 因为其他的读不影响其他的写和读.







# 继续 lib\time文件夹


  update.bash 更新zoneinfo.zip的数据.
  mkzip.go是自己实现的压缩工具.

  ```
	var zb bytes.Buffer
	zw := zip.NewWriter(&zb) //点开这个zip.NewWriter发现他的参数需要一个io.Writer,再点进去发现是一个接口,接口有一个方法Write(p []byte) (n int, err error), byte.Buffer就是一个实现了write方法的类.所以可以传入.//下面就是写入数据即可.
  w, err := zw.CreateRaw(&zip.FileHeader{
  Name:               name,
  Method:             zip.Store,
  CompressedSize64:   uint64(len(data)),
  UncompressedSize64: uint64(len(data)),
  CRC32:              crc32.ChecksumIEEE(data),
})
  if _, err := w.Write(data); err != nil {
    log.Fatal(err)
  }
  ```






# misc\cgo 这里面提供了很多demo代码


fib.go贴到自己的main.go里面. 把big 里面库包改成"math/big"
```go
package main

import (
	"runtime"

	big "math/big"
)

func fibber(c chan *big.Int, out chan string, n int64) {
	// Keep the fibbers in dedicated operating system
	// threads, so that this program tests coordination
	// between pthreads and not just goroutines.
	runtime.LockOSThread() //测试这个在各个os的pthread上的性能.

	i := big.NewInt(n)
	if n == 0 {
		c <- i
	}
	for {
		j := <-c
		out <- j.String() //这里之所以使用string化.是因为big.int直接打印会打印一个对象的地址.不方便观看.转int存不下,只能string化打印是最好的方法.
		i.Add(i, j)   // i=i+j
		c <- i
	}
}

func main() {
	c := make(chan *big.Int)
	out := make(chan string)
	go fibber(c, out, 0) 
	go fibber(c, out, 1)  // 最后参数0,1只是初始值, 线程起来之后, 两个进程就一直死循环了.一直是c里面塞进去一个数, 一个进程读走,之后加上后再塞入. 所以本质还是单进程. 两个进程通过c来同步.效率不会比单进程快.
	for i := 0; i < 200; i++ {
		println(<-out)
	}
}

```


gmp.go 是一个go中嵌入c函数的例子.不好编译,这里就跳过了.
整体思路跟go的bigInt类似.


misc\cgo\gmp\pi.go  使用gmp.go里面的大整数

misc里面其他的都是一些其他平台的支持工具.

## 下面开始源码部分.

src里面从依赖最少得开始看:
	从unicode文件夹开始.文件大体分析会写这里,代码细节我会直接加到相关代码里面的注释.


# src\unicode\utf8\utf8.go

	utf8是unicode一种. 用4个8位来表示. 我理解是我们经常用16进制表示.所以32位=4个16进制.
	对于这个源码我们直接看他提供的接口函数.
	utf8有2个表示一个是byte[] 一个是rune. 

# src\unicode\utf16\utf16.go
	很类似, 这次使用8个8位表示. 如果16进制数,就是4个16进制的数来表示一个unicode编码.


# src\unicode\casetables.go
	定义了一个大小写转化
# src\unicode\graphic.go
	图形的unicode字符串.

# src\unicode\letter.go
	一些字符串的转化函数. 大小写, 在不在一个范围,属性啥的工具函数.

# src\unicode\tables.go
	一些常量的表,作为数据用.不用分析里面的代码逻辑.



# src\unsafe\unsafe.go
	这个库包绕过了go的类型检查,所以不安全.可以直接访问变量的内存和指针.所以很方便.代码中只有函数名和大量的注释.所以这里把注释进行了一些翻译.估计这些函数实现的代码在其他部分.



# src\strings\builder.go
		builder是用来创建字符串的.
# src\strings\clone.go
# src\strings\compare.go
# src\strings\reader.go
	 提供了读取字符串数据的各个函数,也都比较简单.
# src\strings\replace.go
	 使用trie树来进行批量的线程安全的字符串替换工具.


# src\strings\search.go
	 bm算法的实现.用于批量的替换字符串.每一次替换很多组字符串对.
	 bm算法来找一个字符串的子串匹配. 这个算法比kmp一般快3倍.

# src\strings\strings.go
	 一些字符串基本操作

# strconv
	 字符串的转化
	 # src\strconv\atob.go  ascii到bool的转化
	 # src\strconv\atoc.go  ascii到复数
	 # src\strconv\decimal.go 小数的实现.不建议深入研究,因为这个对于float十进制的不是精确的.实用性不高.
	 # src\strconv\ftoa.go 浮点数到字符串
	 # src\strconv\atof.go 字符串到float
	 # src\strconv\atoi.go 字符串到int
	 # src\strconv\bytealg.go  字符串index函数.
			src\strconv\bytealg.go:13行 引用的是 src\internal\bytealg\indexbyte_native.go:13行
			实际实现在src\internal\bytealg\indexbyte_generic.go
			src\internal\bytealg\indexbyte_generic.go:9 里面写了如何用这个go生成各个平台的汇编代码.
			汇编会涉及一些类型的底层实现:
			type slice struct {//切片
				array unsafe.Pointer
				len int
				cap int
			}
			一个[]byte​ 自然也没有什么特殊的， 也是这样的一个slice​结构， 其中的array​指向一个byte array​。
			type strStruct struct {//string
				str unsafe.Pointer
				len int
			}
			这样我们就可以解释汇编代码了.
			```
TEXT	·IndexByte(SB), NOSPLIT, $0-40
	MOVQ b_base+0(FP), SI   //b_base是一个Pointer所以占用8位
	MOVQ b_len+8(FP), BX   //这里其实是b_len和b_cap两个int,所以占用16位.
	MOVB c+24(FP), AL      //c是int所以占用8位.
	LEAQ ret+32(FP), R8         
	JMP  indexbytebody<>(SB)
			
			```
# src\strconv\isprint.go
	一些编码是否可以打印编码有32位和16位的

# src\strconv\itoa.go
  整数到字符串




























		# math
		浮点数基本资料:https://blog.csdn.net/weixin_47713503/article/details/108699001
		这里面我们需要记住几个关键数值:在下面一些代码中有用到.
			长浮点数的各个位: 符号位1, 阶码11, 尾数码52, 总位数64, 偏置值3FFH, 十进制偏置值1023

		里面有大量的汇编.文件结构是函数名_平台.s.里面很多函数都涉及数学上的算法.
		我们只需要看amd64或者x86的即可.这俩是pc平台.如果不写平台的就是跨平台的,是必看的.

		
		先看math里面根目录的代码, 他们依赖最少.都是一些数学运算不涉及过多代码设计.
		最底层是

		# src\math\unsafe.go 
		这个代码通过unsafe的指针转化来进行float32 跟bits float64跟bits的互化.

		# src\math\abs.go
		# src\math\acosh.go
		里面有函数避免整数溢出的优化策略

		# src\math\const.go
			一些常数. pi, E啊啥的, 都是多少位的近似值.
		# src\math\atan.go
			利用多项式近似计算.
		# src\math\asin.go
			利用atan来计算.
		# src\math\atan2.go
		  也是利用atan
		# src\math\atanh.go
		  里面也是用了分段函数,来保证计算的最大程度精确.
		# src\math\bits.go
		  是一些特殊数值的定义.无穷,负无穷等和太小的数归一化的方法.
		# src\math\cbrt.go
		  三次开根号算法.牛顿法


		# src\math\dim.go
			一些比较大小函数.
		# src\math\erf.go    src\math\erfinv.go
		  函数erf(x)在数学中为误差函数（也称之为高斯误差函数，error function or Gauss error function），是一个非基本函数（即不是初等函数），其在概率论、统计学以及偏微分方程和半导体物理中都有广泛的应用。是比较高级的函数,涉及一些复杂算法, 这里就不展开了.
		# src\math\exp.go src\math\expm1.go
		  exp函数,底层也是用多项式来近似.
			对于我们amd64平台底层实现是用汇编来加速:(先读懂exp.go的代码,然后再读相关的汇编代码,算法思路都是一样的只是实现的语言不同,并且exp.go里面注释给了算法说明.汇编代码里面没有算法说明.)
			src\math\exp_amd64.s
			使用taylor展开式来近似计算.这个文件的go接口原型在src\math\exp_asm.go


		# src\math\fma.go
		  加速版本的 求x*y+z, 底层思想是用位运算加速, 具体细节比较复杂.
		# src\math\frexp.go
			对一个浮点数进行2的次幂拆分: f == frac × 2**exp,也是用位运算加速.
		# src\math\log.go
			83行看到,如果平台支持,那么就使用汇编来计算log
			根据我们的平台,底层实现是src\math\log_amd64.s

		# src\math\modf.go
			mod拆分一个浮点数为一个整数跟一个分数的和.
		# src\math\nextafter.go
		  返回x到y这个方向的float数的下一个.
		# src\math\pow.go
		  算x的y次幂.使用数学上的换底公式,换成exp和log函数来算.
		# src\math\remainder.go
			x REM y  =  x - [x/y]*y  
		下面是一些二级库包
		# src\math\bits\bits.go
			一些bytes的操作. 加减乘除mod,多少位是1,多少位是0等.
		# src\math\cmplx
			都是一些复数计算, 很少用到.
		# src\math\big\arith.go
			首先复习大端编码,小端编码
			// 1）Little-endian：将低序字节存储在起始地址（低位编址）
			// 2）Big-endian：将高序字节存储在起始地址（高位编址）
			// 记忆: 关注地址的开始地址存什么, 开始存高bit, 就叫大字节序. endian:是end单词加一个后缀表示字节的顺序. 高bit,表示的是大的数. 比如bin(11)里面第一个1表示2, 第二个1表示1.所以就记住了.地址上来就记录大数就是大endian, 地址上来记录小数就是小endian.
			// 如果我们将0x1234abcd写入到以0x0000开始的内存中，则结果为；
			// address	big-endian	little-endian
			// 0x0000		0x12				0xcd
			// 0x0001		0x34				0xab
			// 0x0002		0xab				0x34
			// 0x0003		0xcd				0x12


		# src\math\big\nat.go  src\math\big\natconv.go   src\math\big\natdiv.go
			natural numbers 的方法和定义.他是大整数, 大有理数, 大浮点数的底层.
		# src\math\big\int.go 大整数
		# src\math\big\float.go 浮点数 表示为:sign × mantissa × 2**exponent
		# src\math\big\rat.go 分数
		# src\math\big\sqrt.go 算根号. 牛顿法.
			使用sync.Once来保证全局变量的初始化唯一一次.节省内存资源.
		下面的rand库包都是计算分布函数里面的抽样.

		# src\math\rand\rng.go
			是均匀分布的实现.核心是0到2^32次幂区间的均匀采样的实现
			理解一下里面的计算流程,至于为什么算法这么设计是对的,需要看相关论文.
		# src\math\rand\rand.go
			伪随机数.利用上面的rng.go来实现(0,n)之间的int, float均匀抽样.
		# src\math\rand\exp.go
			这个是计算指数分布里面的抽样.

		# src\math\rand\zipf.go
		  zipf分布的采样, 都比较简单.
		# src\math\rand\normal.go 
			都是空间放缩, 细节参考注释内的论文.
		# src\math\rand\exp.go   同上
		# src\math\rand\v2 里面内容不太常用.
			
		# src\maps\maps.go
			提供了map对象的equal方法.
			~int, ~string 等各种类型前添加一个波浪线 ~，表示的是衍生类型，即使用 type 自定义的类型也可以被识别到(type MyInt int)，底层类型一致即可。
			实现都比较简单.

		#src\time\tzdata\tzdata.go
			go可以直接函数声明,不写函数实现.
		# src\time\format.go
			时间和字符串的转化.
		# src\time\time.go
			时间对象和方法.不难看懂. 记住time的结构体比较有用. 他结构体是记录纳秒,和一个时区的信息.
		# src\time\zoneinfo.go
		  实现了时区的信息.
		# src\time\sleep.go
			使用channel来实现timer计时器.
		# src\time\tick.go
		  跟上面sleep类似.
		# src\time\genzabbrs.go
			用来生成时区信息.
		# src\time\sys_windows.go
			文件读取的系统函数.
		# src\sync\atomic\asm.s
			具体实现都在runtime里面的汇编.后续再看底层实现.这里面给的是接口的函数原型.
			可以看到里面操作的都是32位或者64位的整数
			这些sync代码可以看到都是nocopy的,只要一个接口实现了lock和unlock,他就是nocopy的.
			但是可以拷贝*mutex.
			至于为什么锁和atomic我们都禁止他深拷贝. 但是这个东西不是强制的, 代码里面你可以复制nocopy的, 但是go vet竞争检测时候会提醒你这么做不安全.
			因为深拷贝的锁,完全是一个新的.只是里面状态跟之前锁一样, 之后的操作(加锁,解锁)跟原来的锁没关系(可以写一个mutex锁复制代码,测试一下看看里面的state如何继承原锁,而后续操作又不继承原锁),那么既然没关系,为何不新创立一个对象锁.所以go里面直接建议禁止copy, 来维护代码的清晰.如果复制锁会让代码非常难以理解.
		# src\sync\atomic\type.go
			对上面asm.s进行的封装.让他可以支持更多的类型的元操作.
		# src\sync\atomic\value.go
			对任意类型的进行元操作支持.
			里面的Store函数是并发的优秀模型.
		# src\sync\map.go
		  并发安全的map模型.
		# src\sync\once.go
			并发限制执行单次模型.
		# src\sync\oncefunc.go
			上面单次模型的拓展,

		# src\sync\cond.go
		  大部分场景下使用 channel 是比 sync.Cond方便的。不过我们要注意到，sync.Cond 提供了 Broadcast 方法，可以通知所有的等待者。想利用 channel 实现这个方法还是不容易的。我想这应该是 sync.Cond 唯一有用武之地的地方。
			这里面的check() 函数是一个经典的nocopy实现,如果不理解以后可以直接复制这段代码用作自己结构体nocopy的实现.如果理解了也可以自己根据自己需要进行改造.

		# src\sync\poolqueue.go
			单生产者,多消费者模型.
		# src\sync\pool.go
			涉及比较多的底层.了解go的GMP模型 https://cloud.tencent.com/developer/article/2409305
		# src\sync\rwmutex.go
		# src\sync\waitgroup.go
		  这些都涉及底层的runtime库包.可以先留着.
		# src\arena\arena.go
			一块内存的同时申请和释放.
		# src\bufio\bufio.go
			reader writer readwriter三个类型,分别里面有一个buf用来维护读写缓存.
		# src\bufio\scan.go
			提供了一个比reader更方便的读取文本的类,可以读取之后,进行分割.
			创建scanner只需要提供一个io.Reader即可.
		# src\bytes\bytes.go
			字符串的一些方法
		# src\bytes\reader.go
		  读写方法
		# src\bytes\buffer.go
			字符串buffer提供读写
		# src\cmp\cmp.go
			比较
		# src\compress\bzip2\move_to_front.go
			移动byte的算法

		# src\compress\bzip2\bit_reader.go
			这里我们要区分计算机存储的概念, 这里面bit 是比特位 是一个01表示.  byte是比特是8位二进制 byte=8bits.这里面读取更底层是按照位来读取的.
		# src\compress\bzip2\bzip2.go
			这是字符串的压缩算法了.
			compress库包里面其他算法实现也都是类似实现编解码.
		# src\container\heap\heap.go
			堆的实现.他的定义是使用接口定义的.
			type Interface interface {
				sort.Interface
				Push(x any) // add x as element Len()
				Pop() any   // remove and return element Len() - 1.
			}
			凡是实现了这3个方法的类都可以把他视作heap.从而heap更广义,比其他语言使用更方便, 比如go里面heap就是依赖数组定义的,go不做这个限制.只需要有sort, push,pop方法的类即可做heap.
		# src\container\list\list.go
			双向链表
		# src\container\ring\ring.go
			双向环形链表.创建时候指定大小,数据是一个圈.
		# src\context\context.go
			cancelCtx是里面的核心类.通过通道来控制上下文的关闭.
		# src\crypto
			加密相关算法
		# src\database\sql\driver\driver.go
			定义数据库连接的接口.
		# src\database\sql\driver\types.go
			go类型和数据库里面数据类型转化.

		# src\embed\embed.go
			//go:embed 使用这个标识可以把一个静态文件变成一个变量.
			静态资源访问没有 io 操作，速度会非常快。
		# src\errors\errors.go
			本质是字符串.
		# src\errors\join.go
			把一组errors拼接成一个长的error字符串
		# src\errors\wrap.go
			errors多维数组里面进行错误匹配,搜索.
		# src\expvar\expvar.go
			提供几个类用于当全局变量.里面的操作都是atomic的.保证并发安全.
		# src\fmt\print.go
			根据format刷新字符串格式,该转换类型的转换类型,转成能打印的字符串,然后交给io处理.

		# src\hash\fnv\fnv.go
			哈希算法, 把字符串看做ascii码的数字,跟一些素数做乘法加法得到哈希值.

		# src\image\color\color.go
			color类
		# src\image\image.go
			图像的image类
		# src\image\color\ycbcr.go
			ycbcr跟rgb的转化
		# src\image\color\palette\gen.go
			调色板的生成
		# src\index\suffixarray\suffixarray.go
			后缀数组用于字符串搜索.


		# src\iter\iter.go
			迭代器:底层在runtime.newcoro
		# src\sort\sort.go
			定义排序算法的接口
		# src\sort\gen_sort_variants.go
			生成文档



#下面是底层部分
		#主要涉及internal, runtime, reflect, go, syscall等目录,都是平时很少用到的go内部调用的底层算法,和汇编,编译器运行时,操作系统相关.


		# src\internal\goarch\gengoarch.go
			生成每一个芯片架构的参数go文件.
		# src\internal\abi\abi_amd64.go
			芯片的寄存器相关参数.
			// RAX, RBX, RCX, RDI, RSI, R8, R9, R10, R11.  //这9个用来存整数.
			IntArgRegs = 9

			// X0 -> X14.
			FloatArgRegs = 15 //这15个用来存float

			// We use SSE2 registers which support 64-bit float operations.  The 8 registers are named xmm0 through xmm7. 这些用来操作8byte的浮点数. float64.
			EffectiveFloatRegSize = 8
		# src\internal\abi\abi.go
			应用二进制接口（英语：application binary interface，缩写为ABI），这块源码的文档是与函数传参以及返回值传递到底是分配在栈还是寄存器上的调用规约
		# src\internal\abi\type.go
			go中数据类型的定义
		# src\internal\abi\compiletype.go
			计算各种类型的占用byte大小.

		# src\internal\bisect\bisect.go
			实现bisect debug工具.

		# src\internal\buildcfg\cfg.go
			利用runtime库返回操作系统信息和硬件信息.

		# src\internal\bytealg\bytealg.go
			RabinKarp字符串搜索子串算法.

		# src\internal\bytealg\compare_generic.go
			里面有byte[] 和string 的字符串比较算法.
		# src\internal\bytealg\compare_amd64.s
			上面go代码编译之后的代码

		# src\internal\bytealg\count_generic.go
			里面有byte[] 和string 的字符串计数算法,都非常简单.

		# src\internal\bytealg\index_amd64.go
			提供index索引算法的支持函数.

		# src\internal\chacha8rand\chacha8_generic.go
			chacha8加密算法.

		# src\internal\coverage
			这个用来提供go test 覆盖率测试的.
		# src\internal\cpu\cpu_x86.go
			提供cpu信息的函数
		# src\internal\cpu\cpu_x86.s
			上面go文件里面的3个函数的汇编源码.因为太常用了,所以汇编来加速.


		# src\internal\dag\alg.go
			dag图的算法.DAG，Directed Acyclic Graph即「有向无环图」。
		# src\internal\dag\parse.go
			dag图的定义和字符串化解析
		# src\internal\diff\diff.go
			比较两个文件的差异, git上使用的算法.
			https://www.jianshu.com/p/fdaeec5dc7ff
			从LCS到IGListKit框架中的Diff算法




		# src\reflect
			属于go的高级用法.读源码之前可以通过go的官方文档.
			整体复习一遍reflect的用法.
			https://go.dev/blog/laws-of-reflection
			这里是这个文档重要部分的记录,有时间的建议阅读上面链接的官方文档.
			1.因为reflet依赖go的类型,所以我们先来看go的类型.
				go是静态类型的. 每一个变量都有一个静态类型,换句话说,只有一个类型并且在编译时候就会固定. 例如: int, float32, *MyType, []byte,等等.(比如python,js这种就是运行时才会决定变量的类型,并且也没编译过程,就是动态类型的语言)
				例如我们定义:
				type MyInt int

				var i int
				var j MyInt
				那么i有类型int, j的类型MyInt,i,j类型不同,他们不能互相赋值,除非进行类型转化.(这是go的设计哲学决定的,保证类型安全)
				一种重要的类型是接口类型. 这种类型绑定了方法.一个接口变量可以保存任意具体的值,只要这些值实现了接口的方法即可.一个非常知名的例子就是io.Reader 和io.Writer.我们看他们的源码:
				// Reader is the interface that wraps the basic Read method.
				type Reader interface {
						Read(p []byte) (n int, err error)
				}

				// Writer is the interface that wraps the basic Write method.
				type Writer interface {
						Write(p []byte) (n int, err error)
				}
			一个变量他是io.Reader类型的,那么他可以保存任意值,只需要这个value的类型有一个Read方法.
			比如:
				var r io.Reader
				r = os.Stdin
				r = bufio.NewReader(r)
				r = new(bytes.Buffer)
				// and so on // 可以看到r的这三种赋值都是正确的.
				非常重要的一点是.不管r具体保存什么类型的值,r的类型始终是io.Reader.Go是静态类型的.r的静态类型就是io.Reader不是bytes.Buffer什么的.
				一个重要例子就是interface{},他可以保存任意类型的值.
				一些人说go的接口是动态类型的,这是错的.因为一个接口他的值在运行时可以任意变化,但是他始终类型就是这个接口类型.这点就引申出来了reflect库包的作用.


				接口的表示:一个接口的变量保存一对信息. 一个是变量具体的值.一个是这个值的类型描述.
				例如:
				var r io.Reader
				tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
				if err != nil {
						return nil, err
				}
				r = tty
				这份代码运行之后. r包含(tty, *os.File),但是r只能使用reader方法.
				注意到我们的类型是*os.File,他包含了超出Read的方法.所以我们可以运行如下代码.
				var w io.Writer
				w = r.(io.Writer)  //可以进行类型转化.
				我们也可以这么干. w只能使用writer方法.
				var empty interface{}
				empty = w
				那么empty底层也是(tty, *os.File).这时empty变量不能有任何方法了.
				这很方便, 因为一个空的interface,包含了全部的value和类型信息.
			The first law of reflection
				1. Reflection goes from interface value to reflection object.
				反射可以让接口的值变成反射对象.
				var x float64 = 3.4
				v := reflect.ValueOf(x)
				fmt.Println("type:", v.Type()) //type: float64
				fmt.Println("kind is float64:", v.Kind() == reflect.Float64)//kind is float64: true
				fmt.Println("value:", v.Float())//value: 3.4
				也就是说valueof的结果再取type,还能得到x的类型. valueof的结果取kind也能得到类型.


				The reflection library has a couple of properties worth singling out. First, to keep the API simple, the “getter” and “setter” methods of Value operate on the largest type that can hold the value: int64 for all the signed integers, for instance. That is, the Int method of Value returns an int64 and the SetInt value takes an int64; it may be necessary to convert to the actual type involved:
				var x uint8 = 'x'
				v := reflect.ValueOf(x)
				fmt.Println("type:", v.Type())                            // uint8.
				fmt.Println("kind is uint8: ", v.Kind() == reflect.Uint8) // true.
				x = uint8(v.Uint())                                       // v.Uint returns a uint64.// uint之后, 会用最大的64来存.这是为了兼容性考虑


				The second property is that the Kind of a reflection object describes the underlying type, not the static type. If a reflection object contains a value of a user-defined integer type, as in

				type MyInt int
				var x MyInt = 7
				v := reflect.ValueOf(x)
				the Kind of v is still reflect.Int, even though the static type of x is MyInt, not int. In other words, the Kind cannot discriminate an int from a MyInt even though the Type can.
			The second law of reflection
			2. Reflection goes from reflection object to interface value.
			反射可以变回到接口
			// Interface returns v's value as an interface{}.
			func (v Value) Interface() interface{}//这里面v是通过reflect.ValueOf函数得到的value.这里面把他变成一个空接口.
			例如:
			y := v.Interface().(float64) // y will have type float64.
			fmt.Println(y)
			to print the float64 value represented by the reflection object v.
			//可以直接写fmt.Println(v)
			Again, there’s no need to type-assert the result of v.Interface() to float64; the empty interface value has the concrete value’s type information inside and Printf will recover it.
			In short, the Interface method is the inverse of the ValueOf function, except that its result is always of static type interface{}.
			The third law of reflection
			3. To modify a reflection object, the value must be settable.
			这条用法是最复杂的,单也是最有意思的.
			Here is some code that does not work, but is worth studying.

			var x float64 = 3.4
			v := reflect.ValueOf(x)
			v.SetFloat(7.1) // Error: will panic.





			If you run this code, it will panic with the cryptic message

			panic: reflect.Value.SetFloat using unaddressable value
			The problem is not that the value 7.1 is not addressable; it’s that v is not settable. Settability is a property of a reflection Value, and not all reflection Values have it.

			The CanSet method of Value reports the settability of a Value; in our case,

			var x float64 = 3.4
			v := reflect.ValueOf(x)
			fmt.Println("settability of v:", v.CanSet())
			prints

			settability of v: false


			When we say

			var x float64 = 3.4
			v := reflect.ValueOf(x)
			we pass a copy of x to reflect.ValueOf

			If we want to modify x by reflection, we must give the reflection library a pointer to the value we want to modify.

			Let’s do that. First we initialize x as usual and then create a reflection value that points to it, called p.//我们来初始化一个x,然后拿一个指针p指向他.

			var x float64 = 3.4
			p := reflect.ValueOf(&x) // Note: take the address of x.//这里我们操作地址.
			fmt.Println("type of p:", p.Type())
			fmt.Println("settability of p:", p.CanSet())
			//The output so far is

			type of p: *float64
			settability of p: false

			我们操作他的指向内容:
			v := p.Elem()//通过Elem来获取指向内容.
			fmt.Println("settability of v:", v.CanSet())//settability of v: true
			//这时我们才可以修改x:
			v.SetFloat(7.1)
			fmt.Println(v.Interface()) //7.1
			fmt.Println(x)             //7.1
			结构体上的应用:
			A common way for this situation to arise is when using reflection to modify the fields of a structure.


			type T struct {
					A int
					B string
			}
			t := T{23, "skidoo"}
			s := reflect.ValueOf(&t).Elem()//s是通过反射得到的真正t底层的本身!,所以下面操作s比t更底层.方法更丰富.我们直接用t不知道他的各个类型,但是用s就知道各个字段类型和值.
			typeOfT := s.Type()//拿到结构体的具体定义类型
			for i := 0; i < s.NumField(); i++ {//遍历s的各个字段. s拿到字段的值,typeOfT拿到各个字段的类型.
					f := s.Field(i)
					fmt.Printf("%d: %s %s = %v\n", i,
							typeOfT.Field(i).Name, f.Type(), f.Interface())
			}


			The output of this program is

			0: A int = 23
			1: B string = skidoo


			Because s contains a settable reflection object, we can modify the fields of the structure.

			s.Field(0).SetInt(77)
			s.Field(1).SetString("Sunset Strip")
			fmt.Println("t is now", t)


			And here’s the result:

			t is now {77 Sunset Strip}
		# 下面我们回到reflect源码
		# src\reflect\value.go
		# src\reflect\type.go
			提供了运行时的value和type的定义.给reflect库使用.
			提供了go基本类型的value和type的一些运行时的方法.用于reflect库里面其他文件的调用.
			这两个类型都过于长,可以挑主要部分看看.

		# src\reflect\visiblefields.go
			返回一个结构体里面的可以访问到的字段
		# src\reflect\swapper.go
			Swapper returns a function that swaps the elements in the provided
			slice.
			给一个slice,然后swapper返回一个函数,这个函数可以交换里面的元素.
		# src\reflect\deepequal.go
			判定两个value是否深度相等. 至于深度相等的定义可以看这个代码里面的注释. 基本可以看做严格相等.


		# src\reflect\asm_amd64.s
			定义了2个函数·makeFuncStub  ·methodValueCall 他俩用在src\reflect\makefunc.go里面.


		# src\reflect\makefunc.go
			MakeFunc把一个函数fn进行了包装,返回一个reflect.Value对象.
			makeMethodValue把reflect.Value包装成makeFuncImpl对象.

		# src\internal\intern\intern.go
			使用一个map的value来把保存任意对象.

		# src\internal\itoa\itoa.go
			int转string(这个itoa里面a表示ascii码数组,所以就是string)


		# src\internal\lazyregexp\lazyre.go
			把re封装了一个once.Do, 保证只有第一次调用任何匹配方法时候才进行compile.
		# src\internal\lazytemplate\lazytemplate.go
			跟上一个go文件同理

		# src\internal\poll\fd_fsync_windows.go
			file syncronize 文件同步为了避免缓存中的数据还没有写入到磁盘就宕机导致的数据丢失，就需要使用fsync或这fdatasync来保证数据成功写入磁盘。

		# src\internal\poll\fd_io_plan9.go
			异步io的读写,可以让另外一个进程来暂停其他进程的异步io读写.
		# src\internal\poll\fd_mutex.go
			文件描述符的锁,主要是读写锁的加锁解锁.
		# src\internal\poll\fd_plan9.go
			封装了fd_mutex的锁的fd类,可以并发读写.
		# src\internal\poll\fd_windows.go
			windows系统下的并发读写.
		# src\internal\poll\fd_poll_runtime.go
			利用runtime库来实现poll功能.


		# 下面我们首先学习go runtime, 他是一个go语言跟system调用的中间层.本着从用户层到内核层的原则,我们先学习runtime库包.因为比较底层,go语言教程一般不学runtime的使用.所以我们先来复习runtime的基本使用方法.
		https://blog.csdn.net/pyf09/article/details/113867102
		里面一些核心概念的总结:
		Golang调度:PMG模型,P是processor.M是multithread.G是goroutine.
		正是因为G足够小,所以切换成本很低.
		线程的运行, 其实是被运行. 其阻塞, 其实是切换出调度队列, 不再去调度执行这个执行流. 其他执行流满足其条件, 便会把被移出调度队列的执行流重新放回调度队列.
		理解了阻塞的真正含义, 也就知道能够比较容易理解, 为什么go的锁, channel这些不阻塞线程. 对于实现的同步执行流效果, 又不阻塞线程的网络


		src\runtime\HACKING.md
		官方的runtime文档


		下面分析runtime源码
		src\runtime\internal\atomic\atomic_amd64.go
		src\runtime\internal\atomic\atomic_amd64.s
			利用汇编的元操作来实现.
		src\runtime\internal\atomic\types.go
			上面的元操作再包一层封装
		src\runtime\internal\math\math.go
			带overflow判断的加法和乘法.

		src\runtime\internal\sys\intrinsics.go
			一些二进制的计算.
		src\runtime\internal\sys\nih.go
			空类型.

		src\runtime\internal\syscall\asm_linux_amd64.s
			系统调用·Syscall6来实现6个参数的linux系统调用的封装.
		src\runtime\internal\syscall\syscall_linux.go
			对于syscall的封装,底层还是syscall6来实现.
		src\runtime\cgo
			提供c/c++跟go的交互,这里不展开.

		src\runtime\asan.go
		src\runtime\asan_amd64.s
		src\runtime\asan0.go
			这种后缀带0的表示系统如果不支持asan技术的话,应该如何写这些处理函数.这里处理方法是跑出异常.
			内存错误分析工具----asan(AddressSanitizer)
			底层函数:getcallerpc,getcallersp由编译器给出.
			// getcallerpc returns the program counter (PC) of its caller's caller.
			// getcallersp returns the stack pointer (SP) of its caller's caller.
			// The implementation may be a compiler intrinsic; there is not
			// necessarily code implementing this on every platform.
			//
			// For example:
			//
			//	func f(arg1, arg2, arg3 int) {
			//		pc := getcallerpc()
			//		sp := getcallersp()
			//	}
			doasanread:是一个底层函数他判断给的addr读取size是否合法,底层用的c语言的__asan_read_go函数.
			其他src\runtime\asan_amd64.s中的函数都跟这个函数类似.

		src\runtime\asm_amd64.s
		提供了runtime里面的汇编代码.包括main函数,哈希算法,go携程实现等.比较长,可以看到其他go代码调用底层时候再找进来读细节比较轻松一点.
		注意这个系统的寄存器和abi寄存器的对应关系.
		// We need to convert to the syscall ABI.
		//
		// arg | ABIInternal | Syscall
		// ---------------------------
		// num | AX          | AX
		// a1  | BX          | DI
		// a2  | CX          | SI
		// a3  | DI          | DX
		// a4  | SI          | R10
		// a5  | R8          | R8
		// a6  | R9          | R9
		//
		// r1  | AX          | AX
		// r2  | BX          | DX
		// err | CX          | part of AX
		https://blog.csdn.net/song_lee/article/details/105297902
		这里面我们知道:
		64 位系统寄存器的作用，64 位程序的前六个参数通过 RDI、RSI、RDX、RCX、R8 和 R9 传递
		所以我们对于代码
		// func memhash(p unsafe.Pointer, h, s uintptr) uintptr
		// hash function using AES hardware instructions
		TEXT runtime·memhash<ABIInternal>(SB),NOSPLIT,$0-32
			// AX = ptr to data
			// BX = seed
			// CX = size
			CMPB	runtime·useAeshash(SB), $0    //比较字节
			JEQ	noaes
			JMP	aeshashbody<>(SB)     //1228行
		noaes:
			JMP	runtime·memhashFallback<ABIInternal>(SB)
		这里面写的AX不是第一个参数而是返回值. 所以对应memhash函数里面的返回值uintptr.
		h是第一个参数所以参考上面表格知道他直接写入bx寄存器, s是第二个参数所以他写入cx寄存器.
		所以我们在memhash中访问bx,cx就直接访问了seed和size. 而AX是数据指针也是返回值的地址.


		src\runtime\string.go
			提供runtime字符串类型,和[]byte,rune等的互化.
		src\runtime\runtime.go
			定义了时间相关的类型.

		src\runtime\runtime2.go
			定义了GMP的类型,字段很多不好理解,等后续这些类型使用的时候,根据方法来解析各个字段的含义用法.	


		src\runtime\lock_sema.go
			锁的信号量实现.






		src\runtime\lockrank.go
			锁的级别
			lockPartialOrder是一个二维数组.定义了各个锁之间的优先级关系.
			比如: lockRankScavenge:        {lockRankSysmon},
			表示lockRankScavenge 获得之前需要先获得lockRankSysmon锁.
		src\runtime\lockrank_on.go

			依赖这个函数:
			// func systemstack(fn func())//这个函数的目的是在系统栈上运行一个给定的函数fn。函数的核心是调用fn之前保存g0到寄存器, 调用函数后把寄存器中的g0再放回g的位置.//g0是一个中转站用来保存g的状态.
			TEXT runtime·systemstack(SB), NOSPLIT, $0-8
				MOVQ	fn+0(FP), DI	// DI = fn  函数地址放到DI上.
				get_tls(CX)     //调用get_tls宏来获取线程局部存储（TLS）的地址，并存储在寄存器CX中。
				MOVQ	g(CX), AX	// AX = g
				MOVQ	g_m(AX), BX	// BX = m
				//下面是3个特例提前退出函数.
				CMPQ	AX, m_gsignal(BX)
				JEQ	noswitch

				MOVQ	m_g0(BX), DX	// DX = g0
				CMPQ	AX, DX
				JEQ	noswitch

				CMPQ	AX, m_curg(BX)
				JNE	bad

				// Switch stacks.
				// The original frame pointer is stored in BP,
				// which is useful for stack unwinding.
				// Save our state in g->sched. Pretend to
				// be systemstack_switch if the G stack is scanned.
				CALL	gosave_systemstack_switch<>(SB)

				// switch to 
				MOVQ	DX, g(CX)      // g0放到g上.
				MOVQ	DX, R14 // set the g register  , g0放R14上
				MOVQ	(g_sched+gobuf_sp)(DX), SP  // g0放sp上

				// call target function
				MOVQ	DI, DX
				MOVQ	0(DI), DI
				CALL	DI

				// switch back to g // 把上面的g0再切回来.
				get_tls(CX)
				MOVQ	g(CX), AX
				MOVQ	g_m(AX), BX
				MOVQ	m_curg(BX), AX
				MOVQ	AX, g(CX)
				MOVQ	(g_sched+gobuf_sp)(AX), SP
				MOVQ	(g_sched+gobuf_bp)(AX), BP
				MOVQ	$0, (g_sched+gobuf_sp)(AX)
				MOVQ	$0, (g_sched+gobuf_bp)(AX)
				RET









		src\runtime\go_tls.h
			go_tls这个宏的实现:
			#define	get_tls(r)	MOVQ TLS, r
			//MOVQ TLS, r: 这是宏展开后的汇编指令。MOVQ是用于在x86-64架构下移动64位数据的指令。TLS通常是一个特殊的段寄存器，它指向当前线程的线程局部存储区域。这个宏的作用是将TLS段寄存器的内容移动到寄存器r中。
			#define	g(r)	0(r)(TLS*1) 获得g指针.

		src\runtime\rwmutex.go
			runtime中使用的读写锁.



# go_read2
# go_read2
