package com.chatserver.client;

// 采用大端字节序
// 尽管java中没有无符号类型，但不影响与服务器正常通信

public class ConvHelper {
	public static int bytesToInt(byte[] b) {
		int i = (b[0] & 0xff) << 24 | (b[1] & 0xff) << 16 | (b[2] & 0xff) << 8 | (b[3] & 0xff);
		return i;
	}
	
	public static byte[] intToBytes(int i) {
		byte[] b = new byte[4];
		b[0] = (byte)(i >> 24);
		b[1] = (byte)(i >> 16);
		b[2] = (byte)(i >> 8);
		b[3] = (byte)(i);
		return b;
	}
	
	public static void main(String[] args) {
		int i = 0x7fffffff;
		byte[] b = intToBytes(i);
		System.out.println(b);
		System.out.println(bytesToInt(b));
		
		i = 123456789;
		b = intToBytes(i);
		System.out.println(b);
		System.out.println(bytesToInt(b));
		
		i = 0;
		b = intToBytes(i);
		System.out.println(b);
		System.out.println(bytesToInt(b));
		
		i = 0xffffffff;
		b = intToBytes(i);
		System.out.println(b);
		System.out.println(bytesToInt(b));
		
	}
}
