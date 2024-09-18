




TEXT ·Sum(SB), $0-8
    MOVQ x+0(FP), AX  // 将第一个参数 x 放入 AX
    MOVQ y+8(FP), BX  // 将第二个参数 y 放入 BX
    ADDQ BX, AX       // 将 BX 加到 AX
    MOVQ AX, ret+16(FP)  // 将结果从 AX 移到返回值位置
    RET               // 返回
