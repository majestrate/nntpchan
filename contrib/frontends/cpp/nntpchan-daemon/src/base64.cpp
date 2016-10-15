#include "base64.hpp"


// taken from i2pd 
namespace i2p
{
namespace data
{
  static void iT64Build(void);

	/*
	*
	* BASE64 Substitution Table
	* -------------------------
	*
	* Direct Substitution Table
	*/

	static const char T64[64] = {
		       'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H',
		       'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P',
		       'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
		       'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f',
		       'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
		       'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
		       'w', 'x', 'y', 'z', '0', '1', '2', '3',
		       '4', '5', '6', '7', '8', '9', '+', '/'
	};

	
	/*
	* Reverse Substitution Table (built in run time)
	*/

	static char iT64[256];
	static int isFirstTime = 1;

	/*
	* Padding 
	*/

	static char P64 = '='; 

	/*
	*
	* ByteStreamToBase64
	* ------------------
	*
	* Converts binary encoded data to BASE64 format.
	*
	*/
	static size_t                                /* Number of bytes in the encoded buffer */
	ByteStreamToBase64 ( 
		    const uint8_t * InBuffer,           /* Input buffer, binary data */
		    size_t    InCount,              /* Number of bytes in the input buffer */ 
		    char  * OutBuffer,          /* output buffer */
		size_t len			   /* length of output buffer */	             
	)

	{
		unsigned char * ps;
		unsigned char * pd;
		unsigned char   acc_1;
		unsigned char   acc_2;
		int             i; 
		int             n; 
		int             m; 
		size_t outCount;

		ps = (unsigned char *)InBuffer;
		n = InCount/3;
		m = InCount%3;
		if (!m)
		     outCount = 4*n;
		else
		     outCount = 4*(n+1);
		if (outCount > len) return 0;
		pd = (unsigned char *)OutBuffer;
		for ( i = 0; i<n; i++ ){
		     acc_1 = *ps++;
		     acc_2 = (acc_1<<4)&0x30; 
		     acc_1 >>= 2;              /* base64 digit #1 */
		     *pd++ = T64[acc_1];
		     acc_1 = *ps++;
		     acc_2 |= acc_1 >> 4;      /* base64 digit #2 */
		     *pd++ = T64[acc_2];
		     acc_1 &= 0x0f;
		     acc_1 <<=2;
		     acc_2 = *ps++;
		     acc_1 |= acc_2>>6;        /* base64 digit #3 */
		     *pd++ = T64[acc_1];
		     acc_2 &= 0x3f;            /* base64 digit #4 */
		     *pd++ = T64[acc_2];
		} 
		if ( m == 1 ){
		     acc_1 = *ps++;
		     acc_2 = (acc_1<<4)&0x3f;  /* base64 digit #2 */
		     acc_1 >>= 2;              /* base64 digit #1 */
		     *pd++ = T64[acc_1];
		     *pd++ = T64[acc_2];
		     *pd++ = P64;
		     *pd++ = P64;

		}
		else if ( m == 2 ){
		     acc_1 = *ps++;
		     acc_2 = (acc_1<<4)&0x3f; 
		     acc_1 >>= 2;              /* base64 digit #1 */
		     *pd++ = T64[acc_1];
		     acc_1 = *ps++;
		     acc_2 |= acc_1 >> 4;      /* base64 digit #2 */
		     *pd++ = T64[acc_2];
		     acc_1 &= 0x0f;
		     acc_1 <<=2;               /* base64 digit #3 */
		     *pd++ = T64[acc_1];
		     *pd++ = P64;
		}
		
		return outCount;
	}

	/*
	*
	* Base64ToByteStream
	* ------------------
	*
	* Converts BASE64 encoded data to binary format. If input buffer is
	* not properly padded, buffer of negative length is returned
	*
	*/
  static
	ssize_t                              /* Number of output bytes */
	Base64ToByteStream ( 
		      const char * InBuffer,           /* BASE64 encoded buffer */
		      size_t    InCount,          /* Number of input bytes */
		      uint8_t  * OutBuffer,	/* output buffer length */ 	
		  size_t len         	/* length of output buffer */
	)
	{
		unsigned char * ps;
		unsigned char * pd;
		unsigned char   acc_1;
		unsigned char   acc_2;
		int             i; 
		int             n; 
		int             m; 
		size_t outCount;

		if (isFirstTime) iT64Build();
		n = InCount/4;
		m = InCount%4;
		if (InCount && !m) 
		     outCount = 3*n;
		else {
		     outCount = 0;
		     return 0;
		}
		
		ps = (unsigned char *)(InBuffer + InCount - 1);
		while ( *ps-- == P64 ) outCount--;
		ps = (unsigned char *)InBuffer;
		
		if (outCount > len) return -1;
		pd = OutBuffer;
		auto endOfOutBuffer = OutBuffer + outCount;		
		for ( i = 0; i < n; i++ ){
		     acc_1 = iT64[*ps++];
		     acc_2 = iT64[*ps++];
		     acc_1 <<= 2;
		     acc_1 |= acc_2>>4;
		     *pd++  = acc_1;
			 if (pd >= endOfOutBuffer) break;

		     acc_2 <<= 4;
		     acc_1 = iT64[*ps++];
		     acc_2 |= acc_1 >> 2;
		     *pd++ = acc_2;
			  if (pd >= endOfOutBuffer) break;	

		     acc_2 = iT64[*ps++];
		     acc_2 |= acc_1 << 6;
		     *pd++ = acc_2;
		}

		return outCount;
	}

	static size_t Base64EncodingBufferSize (const size_t input_size) 
	{
		auto d = div (input_size, 3);
		if (d.rem) d.quot++;
		return 4*d.quot;
	}
	
	/*
	*
	* iT64
	* ----
	* Reverse table builder. P64 character is replaced with 0
	*
	*
	*/

	static void iT64Build()
	{
		int  i;
		isFirstTime = 0;
		for ( i=0; i<256; i++ ) iT64[i] = -1;
		for ( i=0; i<64; i++ ) iT64[(int)T64[i]] = i;
		iT64[(int)P64] = 0;
	}


}
}

namespace nntpchan
{
  std::string B64Encode(const uint8_t * data, const std::size_t l)
  {
    std::string out;
    out.resize(i2p::data::Base64EncodingBufferSize(l));
    i2p::data::ByteStreamToBase64(data, l, &out[0], out.size());
    return out;
  }

  bool B64Decode(const std::string & data, std::vector<uint8_t> & out)
  {
    out.resize(data.size());
    if(i2p::data::Base64ToByteStream(data.c_str(), data.size(), &out[0], out.size()) == -1) return false;
    out.shrink_to_fit();
    return true;
  }
}
