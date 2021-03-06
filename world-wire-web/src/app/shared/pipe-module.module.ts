import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { EscapeHtmlPipe } from './pipes/keep-html.pipe';

@NgModule({
  declarations: [
    EscapeHtmlPipe
  ],
  imports: [
    CommonModule
  ],
  exports: [
    EscapeHtmlPipe
  ]
})
export class PipeModule { }
