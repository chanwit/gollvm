// Taken from: http://npcontemplation.blogspot.com/2008/06/secret-of-llvm-c-bindings.html
package main

/*
#cgo CFLAGS: -D__STDC_LIMIT_MACROS -D__STDC_CONSTANT_MACROS
#cgo LDFLAGS: -lm -lpthread -lpsapi -lstdc++ -lLLVM-3.3

#include <llvm-c/ExecutionEngine.h>
#include <stdio.h>

int RunMe(void* engine, void* fac) {
    LLVMGenericValueRef exec_args[] = {LLVMCreateGenericValueOfInt(LLVMInt32Type(), 10, 0)};
    int p1 = LLVMGenericValueToInt(exec_args[0], 0);
    printf("%d\n", p1);
    // LLVMGenericValueRef exec_res =
    LLVMRunFunction((LLVMExecutionEngineRef)engine, (LLVMValueRef)fac, 1, exec_args);
    // printf("%p\n", exec_res);
    // int result = LLVMGenericValueToInt(exec_res, 0);
    // printf("%d\n", result);
    return 0; // result;
}

*/
import "C"
import "fmt"
// import "unsafe"
import "github.com/chanwit/gollvm/llvm"

func test() {
	llvm.LinkInJIT()
	llvm.InitializeNativeTarget()

    llvm.StartMultithreaded()
    defer llvm.StopMultithreaded()

	mod := llvm.NewModule("fac_module")

	// don't do that, because ExecutionEngine takes ownership over module
	//defer mod.Dispose()

	fac_args := []llvm.Type{llvm.Int32Type()}
	fac_type := llvm.FunctionType(llvm.Int32Type(), fac_args, false)
	fac := llvm.AddFunction(mod, "fac", fac_type)
	fac.SetFunctionCallConv(llvm.CCallConv)
	n := fac.Param(0)
    n.SetName("n")

	entry := llvm.AddBasicBlock(fac, "entry")
	iftrue := llvm.AddBasicBlock(fac, "iftrue")
	iffalse := llvm.AddBasicBlock(fac, "iffalse")
	end := llvm.AddBasicBlock(fac, "end")

	builder := llvm.NewBuilder()
	defer builder.Dispose()

    fmt.Println("DEBUG >>>>")

	builder.SetInsertPointAtEnd(entry);
    fmt.Println("DEBUG 2 >>>>")
	If := builder.CreateICmp(llvm.IntEQ, n, llvm.ConstIntFromString(llvm.Int32Type(), "0", 0), "cmptmp")
    fmt.Println("DEBUG 3 >>>>")
	builder.CreateCondBr(If, iftrue, iffalse)

	builder.SetInsertPointAtEnd(iftrue)
	res_iftrue := llvm.ConstIntFromString(llvm.Int32Type(), "1", 0)
	builder.CreateBr(end)

	builder.SetInsertPointAtEnd(iffalse)
	n_minus := builder.CreateSub(n, llvm.ConstIntFromString(llvm.Int32Type(), "1", 0), "subtmp")
    fmt.Println("DEBUG 4 >>>>")
	call_fac_args := []llvm.Value{n_minus}
	call_fac := builder.CreateCall(fac, call_fac_args, "calltmp")
	res_iffalse := builder.CreateMul(n, call_fac, "multmp")
	builder.CreateBr(end)

	builder.SetInsertPointAtEnd(end)
	res := builder.CreatePHI(llvm.Int32Type(), "result")
	phi_vals := []llvm.Value{res_iftrue, res_iffalse}
	phi_blocks := []llvm.BasicBlock{iftrue, iffalse}
	res.AddIncoming(phi_vals, phi_blocks)
	builder.CreateRet(res)

	err := llvm.VerifyModule(mod, llvm.AbortProcessAction)
	if err != nil {
		fmt.Println(err)
		return
	}

	engine, err := llvm.NewJITCompiler(mod, 2)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer engine.Dispose()

	pass := llvm.NewPassManager()
	defer pass.Dispose()

	pass.Add(engine.TargetData())
	pass.AddConstantPropagationPass()
	pass.AddInstructionCombiningPass()
	pass.AddPromoteMemoryToRegisterPass()
	pass.AddGVNPass()
	pass.AddCFGSimplificationPass()
	pass.Run(mod)

	mod.Dump()
    fmt.Println("DEBUG 5 >>>>")
    gv := llvm.NewGenericValueFromInt(llvm.Int32Type(), 10, true)
    // gv := llvm.ConstIntFromString(llvm.Int32Type(), "1", 0)
    // gv := llvm.ConstIntFromString(llvm.Int32Type(), "10", 1)
    // i := 10;

    // gvref := (*C.LLVMGenericValueRef)(unsafe.Pointer(&GenericVal{0, 10})) //C.LLVMCreateGenericValueOfInt(llvm.Int32Type().C, C.ulonglong(10), C.LLVMBool(0))
    d := C.LLVMGenericValueToInt(gv.C, 1)
    fmt.Printf("DEBUG 6 >>> %d\n", d)

	// exec_args := []llvm.GenericValue{gv}

	// exec_res := engine.RunFunction(fac, exec_args)

    //_ = C.LLVMRunFunction(engine.C, fac.C, 1, gvref)//(*C.LLVMGenericValueRef)(&gvc))

    // i := C.RunMe(unsafe.Pointer(engine.C), unsafe.Pointer(fac.C))

	fmt.Println("-----------------------------------------")
	fmt.Println("Running fac(10) with JIT...")
	// fmt.Printf("Result: %d\n", i)
}

func main() {
	test()
	fmt.Println("DONE")
}
