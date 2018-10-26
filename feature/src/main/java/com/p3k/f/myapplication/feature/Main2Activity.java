package com.p3k.f.myapplication.feature;

import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;

import com.github.barteksc.pdfviewer.PDFView;

public class Main2Activity extends AppCompatActivity {

    PDFView pdfView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main2);
        pdfView = (PDFView) findViewById(R.id.pdfView);
        pdfView.fromAsset("TindakanPertama.pdf").load();
    }
}
