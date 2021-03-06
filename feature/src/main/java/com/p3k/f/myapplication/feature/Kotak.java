package com.p3k.f.myapplication.feature;

import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;

import com.github.barteksc.pdfviewer.PDFView;

public class Kotak extends AppCompatActivity {

    PDFView pdfView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_kotak);

        pdfView = (PDFView) findViewById(R.id.pdfView);
        pdfView.fromAsset("p3k.pdf").load();
    }
}
