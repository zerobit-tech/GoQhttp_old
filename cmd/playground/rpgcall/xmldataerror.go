package main

var xmlDataError string = `<?xml version="1.0"?>
<xmlservice>
    <pgm error="off" lib="sumitg1" name="QHTTPTST81" var="QHTTPTST81">
        <error><![CDATA[*** error sumitg1 QHTTPTST81 ]]></error>
        <version>XML Toolkit 2.0.2-dev</version>
        <error>
            <errnoxml>1100017</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in data]]></xmlerrmsg>
            <xmlhint><![CDATA[p(174)  <data type="5s0" var="inum">         {ZONED}   </dat]]></xmlhint>
        </error>
        <error>
            <errnoxml>1100007</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in excp]]></xmlerrmsg>
            <xmlhint><![CDATA[p(202) <data type="5s0" var="inum"]]></xmlhint>
        </error>
        <error>
            <errnoxml>1100017</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in data]]></xmlerrmsg>
            <xmlhint><![CDATA[p(202) {ZONED}]]></xmlhint>
        </error>
        <error>
            <errnoxml>1100017</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in data]]></xmlerrmsg>
            <xmlhint><![CDATA[p(202) D9C9C361C3C8C1D9]]></xmlhint>
        </error>
        <error>
            <errnoxml>1100007</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in excp]]></xmlerrmsg>
            <xmlhint><![CDATA[p(174) <parm io="both" var="inum"]]></xmlhint>
        </error>
        <error>
            <errnoxml>1100017</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in data]]></xmlerrmsg>
            <xmlhint><![CDATA[p(174)  <data type="5s0" var="inum">         {ZONED}   </dat]]></xmlhint>
        </error>
        <error>
            <errnoxml>1100008</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in failed]]></xmlerrmsg>
            <xmlhint><![CDATA[<pgm error="off" lib="sumitg1" name="QHTTPTST81" var="QHTTP]]></xmlhint>
        </error>
        <error>
            <errnoxml>1100007</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in excp]]></xmlerrmsg>
            <xmlhint><![CDATA[p(174) <parm io="both" var="inum"]]></xmlhint>
        </error>
        <error>
            <errnoxml>1100017</errnoxml>
            <xmlerrmsg><![CDATA[XML copy in data]]></xmlerrmsg>
            <xmlhint><![CDATA[p(174)  <data type="5s0" var="inum">         {ZONED}   </dat]]></xmlhint>
        </error>
        <jobinfo>
            <jobipc>*na</jobipc>
            <jobipcskey>FFFFFFFF</jobipcskey>
            <jobname>QZDASOINIT</jobname>
            <jobuser>QUSER</jobuser>
            <jobnbr>018967</jobnbr>
            <jobsts>*ACTIVE</jobsts>
            <curuser>SUMITG</curuser>
            <ccsid>273</ccsid>
            <dftccsid>273</dftccsid>
            <paseccsid>819</paseccsid>
            <langid>ENG</langid>
            <cntryid>US</cntryid>
            <sbsname>QUSRWRK</sbsname>
            <sbslib>QSYS</sbslib>
            <curlib>SUMITG1</curlib>
            <syslibl>QSYS QSYS2 QHLPSYS QUSRSYS</syslibl>
            <usrlibl>QGPL QTEMP</usrlibl>
            <jobcpffind>see log scan, not error list</jobcpffind>
        </jobinfo>
        <joblogscan>
            <joblogrec>
                <jobcpf>CPF1124</jobcpf>
                <jobtime><![CDATA[23-09-24  20:47:15.182928]]></jobtime>
                <jobtext><![CDATA[Job 018967/QUSER/QZDASOINIT started on 23-09-24 at SUMITG QZBSSECR QzbsChangeJob__Fi 33 Printer device PRT01 not found.]]></jobtext>
            </joblogrec>
            <joblogrec>
                <jobcpf>CPF1301</jobcpf>
                <jobtime><![CDATA[23-09-24  20:47:15.182928]]></jobtime>
                <jobtext><![CDATA[SUMITG QZBSSECR QzbsChangeJob__Fi 33 ACGDTA for 018967/QUSER/QZDASOINIT not journaled; reason Job resource accounting data for j]]></jobtext>
            </joblogrec>
            <joblogrec>
                <jobcpf>SQL799C</jobcpf>
                <jobtime><![CDATA[23-09-24  20:47:15.182928]]></jobtime>
                <jobtext><![CDATA[SUMITG QZDACMDP CP_SET_ATTRCODE 22089 QZDACMDP CP_SET_ATTRCODE 22089 The following special registers have been set: This message]]></jobtext>
            </joblogrec>
            <joblogrec>
                <jobcpf>RNX0105</jobcpf>
                <jobtime><![CDATA[23-09-24  20:47:15.182928]]></jobtime>
                <jobtext><![CDATA[SUMITG QRNXMSG SignalException 22 PLUGILE CPYINDEC 5478 A character representation of a numeric value is in error.]]></jobtext>
            </joblogrec>
        </joblogscan>
        <joblog job='QZDASOINIT' user='QUSER' nbr='018967'> XXXX</joblog>
    </pgm>
</xmlservice>`
