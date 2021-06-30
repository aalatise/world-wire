import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FlexLayoutModule } from '@angular/flex-layout';
import { DropdownModule, ModalModule, LoadingModule, ButtonModule, InputModule } from 'carbon-components-angular';

import { NodeSelectComponent } from './components/node-select/node-select.component';
import { QuickFilterComponent } from './components/quick-filter/quick-filter.component';
import { AccountModalComponent } from './components/account-modal/account-modal.component';

@NgModule({
  declarations: [
    NodeSelectComponent,
    QuickFilterComponent,
    AccountModalComponent,
  ],
  exports: [
    NodeSelectComponent,
    QuickFilterComponent,
  ],
  imports: [
    FormsModule,
    FlexLayoutModule,
    DropdownModule,
    ModalModule,
    ButtonModule,
    InputModule,
    LoadingModule,
    CommonModule
  ],
})
export class PortalSharedModule { }
